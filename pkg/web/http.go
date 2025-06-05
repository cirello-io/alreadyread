// Copyright 2023 cirello.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web // import "cirello.io/alreadyread/pkg/web"

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"

	"cirello.io/alreadyread/frontend"
	"cirello.io/alreadyread/pkg/bookmarks"
	"github.com/rs/cors"
)

//go:generate moq -out urltitleloader_mocks_test.go . URLTitleLoader
type URLTitleLoader interface {
	Title(url string) string
}

// Server implements the web interface.
type Server struct {
	bookmarks *bookmarks.Bookmarks
	cors      *cors.Cors
	handler   http.Handler

	titleLoader URLTitleLoader
}

// New creates a web interface handler.
func New(bookmarks *bookmarks.Bookmarks, titleLoader URLTitleLoader, allowedOrigins []string) *Server {
	s := &Server{
		bookmarks: bookmarks,
		cors: cors.New(cors.Options{
			AllowedOrigins: allowedOrigins,
		}),
		titleLoader: titleLoader,
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	router := http.NewServeMux()

	router.HandleFunc("/post", s.post)
	router.HandleFunc("/inbox", s.inbox)
	router.HandleFunc("/duplicated", s.duplicated)
	router.HandleFunc("/dead", s.dead)
	router.HandleFunc("/all", s.all)
	router.HandleFunc("/search", s.search)
	router.HandleFunc("/bookmarks/", s.bookmarkOperations)
	router.HandleFunc("/", s.index())
	s.handler = s.cors.Handler(router)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Server) post(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	loadTitle := r.URL.Query().Get("loadTitle") == "true"
	bookmark := &bookmarks.Bookmark{
		URL: url,
	}
	if loadTitle {
		bookmark.Title = s.titleLoader.Title(url)
	}
	buf := &bytes.Buffer{}
	frontend.RenderNewLink(buf, bookmark)
	if r.Header.Get("HX-Request") != "true" {
		indexBuf := &bytes.Buffer{}
		frontend.RenderIndex(indexBuf, r.URL.Path, template.HTML(buf.String()))
		buf = indexBuf
	}
	_, _ = io.Copy(w, buf)
}

func (s *Server) inbox(w http.ResponseWriter, r *http.Request) {
	list, err := s.bookmarks.Inbox()
	if err != nil {
		log.Println("cannot load bookmarks for inbox:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	s.renderList(w, r, "Inbox", list)
}

func (s *Server) duplicated(w http.ResponseWriter, r *http.Request) {
	list, err := s.bookmarks.Duplicated()
	if err != nil {
		log.Println("cannot load duplicated bookmarks:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	s.renderList(w, r, "Duplicated", list)
}

func (s *Server) dead(w http.ResponseWriter, r *http.Request) {
	list, err := s.bookmarks.Dead()
	if err != nil {
		log.Println("cannot load dead bookmarks:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	s.renderList(w, r, "Dead", list)
}

func (s *Server) all(w http.ResponseWriter, r *http.Request) {
	list, err := s.bookmarks.All()
	if err != nil {
		log.Println("cannot load all bookmarks:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	s.renderList(w, r, "All", list)
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	list, err := s.bookmarks.Search(r.URL.Query().Get("term"))
	if err != nil {
		log.Println("cannot load all bookmarks:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	s.renderList(w, r, "Search", list)
}

func (s *Server) renderList(w http.ResponseWriter, r *http.Request, title string, list []*bookmarks.Bookmark) {
	buf := &bytes.Buffer{}
	frontend.RenderLinkTable(buf, list)
	if r.Header.Get("HX-Request") != "true" {
		indexBuf := &bytes.Buffer{}
		frontend.RenderIndex(indexBuf, r.URL.Path, template.HTML(buf.String()))
		buf = indexBuf
	} else {
		fmt.Fprintln(buf, "<h2 id=\"header-page-name\" hx-swap-oob=\"true\">", title, "</h2>")
	}
	_, _ = io.Copy(w, buf)
}

func (s *Server) bookmarkOperations(w http.ResponseWriter, r *http.Request) {
	id, err := extractID("/bookmarks", r.URL.String())
	if err != nil {
		log.Println("cannot parse bookmark ID:", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodDelete:
		err := s.bookmarks.DeleteByID(id)
		if err != nil {
			log.Println("cannot delete bookmark:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		return
	case http.MethodPatch:
		if inbox := r.FormValue("inbox"); inbox != "" {
			if err := s.bookmarks.UpdateInbox(id, inbox); err != nil {
				log.Println("cannot update bookmark:", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		return
	case http.MethodPost:
		title, url, description := r.FormValue("title"), r.FormValue("url"), r.FormValue("description")
		if url == "" {
			log.Println("malformed bookmardk")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		err := s.bookmarks.Insert(&bookmarks.Bookmark{
			Title:       title,
			URL:         url,
			Description: description,
		})
		if err != nil {
			log.Println("cannot store new bookmark:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		s.inbox(w, r)
		return
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

}

func (s *Server) index() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			frontend.RenderIndex(w, r.URL.Path, frontend.EmptyContainer)
			return
		}
		http.NotFound(w, r)
	}
}

func extractID(root, urlPath string) (int64, error) {
	if root == "" {
		return 0, errors.New("empty root")
	}
	if urlPath == "" {
		return 0, errors.New("empty URL Path")
	}
	root = path.Clean(root + "/")
	urlPath = path.Clean(urlPath + "/")
	if !strings.HasPrefix(urlPath, root) {
		return 0, errors.New("URL Path doesn't start with root:" + urlPath + " " + root)
	}
	urlPath = strings.TrimPrefix(urlPath, root)
	if urlPath == "" {
		return 0, nil
	}
	urlPath = urlPath[1:]
	urlPathParts := strings.Split(strings.Trim(urlPath, "/"), "/")
	return strconv.ParseInt(urlPathParts[0], 10, 64)
}
