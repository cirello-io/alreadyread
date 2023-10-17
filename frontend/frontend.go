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
	"embed"
	"html/template"
	"io"
	"net/http"

	"cirello.io/alreadyread/pkg/bookmarks"
)

//go:embed assets
var Content embed.FS

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
	linkTable    = template.Must(template.New("linkTable").Parse(linkTableTPL))
)

func RenderLinkTable(w io.Writer, list []*bookmarks.Bookmark) {
	if err := linkTable.Execute(w, list); err != nil {
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

func RenderIndex(w io.Writer, container template.HTML) {
	if err := index.Execute(w, struct{ Container template.HTML }{container}); err != nil {
		if rw, ok := w.(http.ResponseWriter); ok {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}
