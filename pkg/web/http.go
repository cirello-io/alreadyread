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
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"path"
	"slices"
	"strconv"
	"strings"

	"cirello.io/alreadyread/frontend"
	"cirello.io/alreadyread/pkg/actions"
	"cirello.io/alreadyread/pkg/models"
	"github.com/jmoiron/sqlx"
)

// Server implements the web interface.
type Server struct {
	db      *sqlx.DB
	handler http.Handler
}

// New creates a web interface handler.
func New(db *sqlx.DB) *Server {
	s := &Server{
		db: db,
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	rootHandler := http.FileServer(http.FS(frontend.Content))
	router := http.NewServeMux()

	// new
	router.HandleFunc("/newBookmark", s.newBookmark)
	router.HandleFunc("/bookmarks", s.bookmarks)
	router.HandleFunc("/bookmarks/", s.bookmarkOperations)

	router.HandleFunc("/", rootHandler.ServeHTTP)
	s.handler = router
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Server) newBookmark(w http.ResponseWriter, r *http.Request) {
	// TODO: handle Access-Control-Allow-Origin correctly
	w.Header().Set("Access-Control-Allow-Origin", "*")

	bookmark := &models.Bookmark{}
	if err := json.NewDecoder(r.Body).Decode(bookmark); err != nil {
		log.Println("cannot unmarshal bookmark request:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	if err := actions.AddBookmark(s.db, bookmark); err != nil {
		log.Println("cannot save bookmark:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Write([]byte("{}"))
}

func (s *Server) bookmarks(w http.ResponseWriter, r *http.Request) {
	// TODO: handle Access-Control-Allow-Origin correctly
	w.Header().Set("Access-Control-Allow-Origin", "*")

	bookmarks, err := actions.ListBookmarks(s.db)
	if err != nil {
		log.Println("cannot load all bookmarks:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	switch r.URL.Query().Get("filter") {
	case "new":
		bookmarks = slices.DeleteFunc(bookmarks, func(bookmark *models.Bookmark) bool {
			return bookmark.Inbox != models.New
		})
	case "duplicated":
		duplicates := map[string][]*models.Bookmark{}
		for _, bookmark := range bookmarks {
			duplicates[bookmark.URL] = append(duplicates[bookmark.URL], bookmark)
		}
		bookmarks = nil
		for _, duplicate := range duplicates {
			if len(duplicate) > 1 {
				bookmarks = append(bookmarks, duplicate...)
			}
		}
		slices.SortFunc(bookmarks, func(a, b *models.Bookmark) int {
			return b.CreatedAt.Compare(a.CreatedAt)
		})
	}

	if err := frontend.LinkTable.Execute(w, bookmarks); err != nil {
		log.Println("cannot render link table: ", err)
	}
}

func (s *Server) bookmarkOperations(w http.ResponseWriter, r *http.Request) {
	// TODO: handle Access-Control-Allow-Origin correctly
	w.Header().Set("Access-Control-Allow-Origin", "*")

	id, err := extractID("/bookmarks", r.URL.String())
	if err != nil {
		log.Println("cannot parse bookmark ID:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		err := actions.DeleteBookmarkByID(s.db, id)
		if err != nil {
			log.Println("cannot delete bookmark:", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		return
	case http.MethodPatch:
		if inbox := r.FormValue("inbox"); inbox != "" {
			err := actions.UpdateInbox(s.db, id, inbox)
			if err != nil {
				log.Println("cannot update bookmark:", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
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
	urlPath = strings.TrimPrefix(urlPath, root)[1:]
	urlPathParts := strings.Split(strings.Trim(urlPath, "/"), "/")
	return strconv.ParseInt(urlPathParts[0], 10, 64)
}
