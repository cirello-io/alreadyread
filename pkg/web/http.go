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
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"

	"cirello.io/alreadyread/frontend"
	"cirello.io/alreadyread/pkg/bookmarks"
)

// Server implements the web interface.
type Server struct {
	allowedOrigins map[string]struct{}
	bookmarks      *bookmarks.Bookmarks
	handler        http.Handler
}

// New creates a web interface handler.
func New(bookmarks *bookmarks.Bookmarks, allowedOrigins []string) *Server {
	s := &Server{
		bookmarks: bookmarks,
	}
	s.allowedOrigins = make(map[string]struct{})
	for _, allowedOrigin := range allowedOrigins {
		s.allowedOrigins[allowedOrigin] = struct{}{}
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	rootHandler := http.FileServer(http.FS(frontend.Content))
	router := http.NewServeMux()

	router.HandleFunc("/newLink", s.newLink)
	router.HandleFunc("/inbox", s.inbox)
	router.HandleFunc("/duplicated", s.duplicated)
	router.HandleFunc("/all", s.all)
	router.HandleFunc("/search", s.search)
	router.HandleFunc("/bookmarks/", s.bookmarkOperations)
	router.HandleFunc("/", rootHandler.ServeHTTP)
	s.handler = router
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.handleCORS(w, r) {
		return
	}
	s.handler.ServeHTTP(w, r)
}

func (s *Server) handleCORS(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if _, ok := s.allowedOrigins[origin]; ok {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return false
	}
	return true
}

func (s *Server) newLink(w http.ResponseWriter, r *http.Request) {
	bookmark := &bookmarks.Bookmark{
		URL: r.URL.Query().Get("url"),
	}
	bookmark = bookmarks.NewURLChecker().Check(bookmark)
	if err := frontend.LinkTable.ExecuteTemplate(w, "newLink", bookmark); err != nil {
		log.Println("cannot render new bookmark form:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) inbox(w http.ResponseWriter, r *http.Request) {
	list, err := s.bookmarks.Inbox()
	if err != nil {
		log.Println("cannot load bookmarks for inbox:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := frontend.LinkTable.Execute(w, list); err != nil {
		log.Println("cannot render link table for inbox: ", err)
	}
}

func (s *Server) duplicated(w http.ResponseWriter, r *http.Request) {
	list, err := s.bookmarks.Duplicated()
	if err != nil {
		log.Println("cannot load duplicated bookmarks:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := frontend.LinkTable.Execute(w, list); err != nil {
		log.Println("cannot render link table for duplicated list: ", err)
	}
}

func (s *Server) all(w http.ResponseWriter, r *http.Request) {
	list, err := s.bookmarks.All()
	if err != nil {
		log.Println("cannot load all bookmarks:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := frontend.LinkTable.Execute(w, list); err != nil {
		log.Println("cannot render link table: ", err)
	}
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	list, err := s.bookmarks.Search(r.URL.Query().Get("term"))
	if err != nil {
		log.Println("cannot load all bookmarks:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := frontend.LinkTable.Execute(w, list); err != nil {
		log.Println("cannot render link table: ", err)
	}
}

func (s *Server) bookmarkOperations(w http.ResponseWriter, r *http.Request) {

	id, err := extractID("/bookmarks", r.URL.String())
	if err != nil {
		log.Println("cannot parse bookmark ID:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
			err := s.bookmarks.UpdateInbox(id, inbox)
			if err != nil {
				log.Println("cannot update bookmark:", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		return
	case http.MethodPost:
		err := s.bookmarks.Insert(&bookmarks.Bookmark{
			Title: r.FormValue("title"),
			URL:   r.FormValue("url"),
		}, bookmarks.NewURLChecker())
		if err != nil {
			log.Println("cannot store new bookmark:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, `
		<div class="alert" role="alert" data-hx-trigger="load delay:5s" data-hx-on::load="javascript: htmx.trigger('#inbox-btn','click',{})">Bookmark saved!</div>
		`)
		return
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
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
