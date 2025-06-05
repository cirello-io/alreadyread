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
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cirello.io/alreadyread/pkg/bookmarks"
)

func TestRenderNewLink(t *testing.T) {
	t.Run("badWriter", func(t *testing.T) {
		brw := &badResponseWriter{}
		RenderNewLink(brw, nil)
		if brw.recordedStatusCode != http.StatusInternalServerError {
			t.Fatal("unexpected status code:", brw.recordedStatusCode)
		}
	})
	t.Run("good", func(t *testing.T) {
		const (
			expectedURL   = "%FIND-URL%"
			expectedTitle = "%FIND-TITLE%"
		)
		rw := httptest.NewRecorder()
		RenderNewLink(rw, &bookmarks.Bookmark{
			ID:    1,
			URL:   expectedURL,
			Title: expectedTitle,
		})
		body := rw.Body.String()
		if !strings.Contains(body, expectedURL) {
			t.Error("cannot find URL pattern")
		}
		if !strings.Contains(body, expectedTitle) {
			t.Error("cannot find title pattern")
		}
	})
}

func TestRenderLinkTable(t *testing.T) {
	t.Run("badWriter", func(t *testing.T) {
		brw := &badResponseWriter{}
		RenderLinkTable(brw, nil)
		if brw.recordedStatusCode != http.StatusInternalServerError {
			t.Fatal("unexpected status code:", brw.recordedStatusCode)
		}
	})
	t.Run("good", func(t *testing.T) {
		const (
			expectedURL    = "%FIND-URL%"
			expectedTitle  = "%FIND-TITLE%"
			expectedReason = "%FIND-REASON%"
		)
		rw := httptest.NewRecorder()
		RenderLinkTable(rw, []*bookmarks.Bookmark{
			{
				ID:               1,
				URL:              expectedURL,
				Title:            expectedTitle,
				LastStatusReason: expectedReason,
			},
		})
		body := rw.Body.String()
		if !strings.Contains(body, expectedURL) {
			t.Error("cannot find URL pattern")
		}
		if !strings.Contains(body, expectedTitle) {
			t.Error("cannot find title pattern")
		}
		if !strings.Contains(body, expectedReason) {
			t.Error("cannot find last status reason pattern")
		}
	})
	t.Run("badStatusCode", func(t *testing.T) {
		const (
			expectedURL   = "%FIND-URL%"
			expectedTitle = "%FIND-TITLE%"
		)
		rw := httptest.NewRecorder()
		RenderLinkTable(rw, []*bookmarks.Bookmark{
			{
				ID:             1,
				URL:            expectedURL,
				Title:          expectedTitle,
				LastStatusCode: http.StatusInternalServerError,
			},
		})
		body := rw.Body.String()
		if !strings.Contains(body, expectedURL) {
			t.Error("cannot find URL pattern")
		}
		if !strings.Contains(body, expectedTitle) {
			t.Error("cannot find title pattern")
		}
		if !strings.Contains(body, http.StatusText(http.StatusInternalServerError)) {
			t.Error("cannot find HTTP status")
		}
	})
}

func TestRenderIndex(t *testing.T) {
	t.Run("badWriter", func(t *testing.T) {
		brw := &badResponseWriter{}
		RenderIndex(brw, "/", "", "")
		if brw.recordedStatusCode != http.StatusInternalServerError {
			t.Fatal("unexpected status code:", brw.recordedStatusCode)
		}
	})
	t.Run("good", func(t *testing.T) {
		const expectedPattern = "%FIND-ME%"
		rw := httptest.NewRecorder()
		RenderIndex(rw, "/", "", expectedPattern)
		body := rw.Body.String()
		if !strings.Contains(body, expectedPattern) {
			t.Error("cannot find pattern")
		}
	})
}

type badResponseWriter struct {
	recordedStatusCode int
}

func (*badResponseWriter) Header() http.Header {
	return http.Header{}
}

func (*badResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("bad write")
}

func (brw *badResponseWriter) WriteHeader(statusCode int) {
	brw.recordedStatusCode = statusCode
}
