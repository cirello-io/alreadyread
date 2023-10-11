// Copyright 2018 github.com/ucirello
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
	"log"
	"net/http"
	"slices"

	"cirello.io/alreadyread/frontend"
	"cirello.io/alreadyread/pkg/actions"
	"cirello.io/alreadyread/pkg/models"
	"cirello.io/alreadyread/pkg/net"
	"cirello.io/alreadyread/pkg/pubsub"
	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
)

// Server implements the web interface.
type Server struct {
	db             *sqlx.DB
	handler        http.Handler
	pubsub         *pubsub.Broker
	authMiddleware *jwtmiddleware.JWTMiddleware
}

// New creates a web interface handler.
func New(db *sqlx.DB, broker *pubsub.Broker) (*Server, error) {
	s := &Server{
		db:     db,
		pubsub: broker,
	}
	err := s.registerRoutes()
	return s, err
}

func (s *Server) registerRoutes() error {
	rootHandler := http.FileServer(http.FS(frontend.Content))
	router := http.NewServeMux()
	// legacy URLs
	router.HandleFunc("/state", s.state)
	router.HandleFunc("/loadBookmark", s.loadBookmark)
	router.HandleFunc("/newBookmark", s.newBookmark)
	router.HandleFunc("/deleteBookmark", s.deleteBookmark)
	router.HandleFunc("/markBookmarkAsRead", s.markBookmarkAsRead)
	router.HandleFunc("/markBookmarkAsPostpone", s.markBookmarkAsPostpone)
	router.HandleFunc("/ws", s.websocket)

	// new
	router.HandleFunc("/bookmarks", s.bookmarks)

	router.HandleFunc("/", rootHandler.ServeHTTP)
	s.handler = router
	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Server) state(w http.ResponseWriter, r *http.Request) {
	// TODO: handle Access-Control-Allow-Origin correctly
	w.Header().Set("Access-Control-Allow-Origin", "*")
	bookmarks, err := actions.ListBookmarks(s.db)
	if err != nil {
		log.Println("cannot load all bookmarks:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(bookmarks); err != nil {
		log.Println("cannot marshal bookmarks:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (s *Server) loadBookmark(w http.ResponseWriter, r *http.Request) {
	// TODO: handle Access-Control-Allow-Origin correctly
	w.Header().Set("Access-Control-Allow-Origin", "*")

	bookmark := &models.Bookmark{}
	if err := json.NewDecoder(r.Body).Decode(bookmark); err != nil {
		log.Println("cannot unmarshal bookmark request:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	bookmark = net.CheckLink(bookmark)

	if err := json.NewEncoder(w).Encode(bookmark); err != nil {
		log.Println("cannot marshal bookmark:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}
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

	if err := actions.AddBookmark(s.db, bookmark, s.pubsub.Broadcast); err != nil {
		log.Println("cannot save bookmark:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Write([]byte("{}"))
}

func (s *Server) deleteBookmark(w http.ResponseWriter, r *http.Request) {
	// TODO: handle Access-Control-Allow-Origin correctly
	w.Header().Set("Access-Control-Allow-Origin", "*")

	bookmark := &models.Bookmark{}
	if err := json.NewDecoder(r.Body).Decode(bookmark); err != nil {
		log.Println("cannot unmarshal bookmark request:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	if err := actions.DeleteBookmark(s.db, bookmark, s.pubsub.Broadcast); err != nil {
		log.Println("cannot save bookmark:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	w.Write([]byte("{}"))
}

func (s *Server) markBookmarkAsRead(w http.ResponseWriter, r *http.Request) {
	// TODO: handle Access-Control-Allow-Origin correctly
	w.Header().Set("Access-Control-Allow-Origin", "*")
	bookmark := &models.Bookmark{}
	if err := json.NewDecoder(r.Body).Decode(bookmark); err != nil {
		log.Println("cannot unmarshal bookmark request:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}
	if err := actions.MarkBookmarkAsRead(s.db, bookmark.ID, s.pubsub.Broadcast); err != nil {
		log.Println("cannot mark bookmark as read:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
}

func (s *Server) markBookmarkAsPostpone(w http.ResponseWriter, r *http.Request) {
	// TODO: handle Access-Control-Allow-Origin correctly
	w.Header().Set("Access-Control-Allow-Origin", "*")
	bookmark := &models.Bookmark{}
	if err := json.NewDecoder(r.Body).Decode(bookmark); err != nil {
		log.Println("cannot unmarshal bookmark request:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}
	if err := actions.MarkBookmarkAsPostpone(s.db, bookmark.ID, s.pubsub.Broadcast); err != nil {
		log.Println("cannot mark bookmark as postpone:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}
	w.Write([]byte("{}"))
}

func (s *Server) websocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("cannot upgrade to websocket:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}
	if ws != nil {
		defer ws.Close()
	}

	unsubscribe := s.pubsub.Subscribe(func(msg interface{}) {
		if err := ws.WriteJSON(msg); err != nil {
			log.Println("cannot write websocket message:", err)
			ws.Close()
		}
	})
	defer unsubscribe()
	defer ws.Close()

	log.Println("listening for pings...")
	for {
		msgType, _, err := ws.NextReader()
		if err != nil || msgType == websocket.CloseMessage {
			return
		}
	}
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
			return bookmark.Inbox == 0
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
