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

package frontend

import (
	_ "embed"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"

	"cirello.io/alreadyread/pkg/bookmarks"
)

var (
	//go:embed newLink.html
	newLinkTPL string
	newLink    = template.Must(template.New("newLink").Parse(newLinkTPL))
)

func RenderNewLink(w io.Writer, bookmark *bookmarks.Bookmark) {
	if err := newLink.Execute(w, bookmark); err != nil {
		if rw, ok := w.(http.ResponseWriter); ok {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

var (
	//go:embed linkTable.html
	linkTableTPL string
	linkTable    = template.Must(template.New("linkTable").Funcs(template.FuncMap{
		"prettyTime":     func(t time.Time) string { return t.Format("Jan _2 2006") },
		"httpStatusCode": func(code int64) string { return http.StatusText(int(code)) },
	}).Parse(linkTableTPL))
)

func RenderLinkTable(w io.Writer, list []*bookmarks.Bookmark) {
	type dateGroup struct {
		Date  string
		Links []*bookmarks.Bookmark
	}
	var (
		idx    = make(map[string]*dateGroup)
		groups = make([]string, 0)
		p      struct {
			Links []*dateGroup
		}
	)
	for _, b := range list {
		date := b.CreatedAt.Format("Jan _2 2006")
		if _, ok := idx[date]; !ok {
			groups = append(groups, date)
			idx[date] = &dateGroup{
				Date:  date,
				Links: []*bookmarks.Bookmark{},
			}
		}
		idx[date].Links = append(idx[date].Links, b)
	}
	for _, g := range groups {
		p.Links = append(p.Links, idx[g])
	}
	if err := linkTable.Execute(w, p); err != nil {
		log.Println("cannot render link table:", err)
		if rw, ok := w.(http.ResponseWriter); ok {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

var (
	//go:embed index.html
	indexTPL string
	index    = template.Must(template.New("index").Parse(indexTPL))
)

const EmptyContainer template.HTML = ""

const NoTitle = ""

func RenderIndex(w io.Writer, path string, headerPageName string, container template.HTML) {
	err := index.Execute(w, struct {
		Path           string
		HeaderPageName string
		Container      template.HTML
	}{Path: path, HeaderPageName: headerPageName, Container: container})
	if err != nil {
		log.Println("cannot render index:", err)
		if rw, ok := w.(http.ResponseWriter); ok {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
}
