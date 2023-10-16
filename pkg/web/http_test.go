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

package web

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cirello.io/alreadyread/pkg/bookmarks"
)

func Test_extractID(t *testing.T) {
	tests := []struct {
		root    string
		url     string
		want    int64
		wantErr bool
	}{
		{"", "/bookmarks/1", 0, true},
		{"/bookmarks", "", 0, true},
		{"/bookmarks", "/bookmarks/1", 1, false},
		{"/bookmarks", "/bookmarks/1/2", 1, false},
		{"/bookmarks", "/bookmarks", 0, false},
		{"/bookmarks", "/banana", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.root+tt.url, func(t *testing.T) {
			got, err := extractID(tt.root, tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer(t *testing.T) {
	t.Run("Inbox", func(t *testing.T) {
		t.Run("badDB", func(t *testing.T) {
			errDB := errors.New("bad DB")
			repository := &RepositoryMock{
				InboxFunc: func() ([]*bookmarks.Bookmark, error) {
					return nil, errDB
				},
			}
			root := bookmarks.New(repository)
			ts := httptest.NewServer(New(root, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/inbox")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Fatal("not OK")
			}
		})
		t.Run("good", func(t *testing.T) {
			foundBookmark := &bookmarks.Bookmark{ID: 1, Title: "%FIND-TITLE%", URL: "https://%FIND-%URL.com"}
			repository := &RepositoryMock{
				InboxFunc: func() ([]*bookmarks.Bookmark, error) {
					return []*bookmarks.Bookmark{
						foundBookmark,
					}, nil
				},
			}
			root := bookmarks.New(repository)
			ts := httptest.NewServer(New(root, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/inbox")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatal("not OK")
			}
			buf := &bytes.Buffer{}
			io.Copy(buf, resp.Body)
			if !strings.Contains(buf.String(), foundBookmark.Title) {
				t.Error("cannot find expected bookmark title")
			}
			if !strings.Contains(buf.String(), foundBookmark.URL) {
				t.Error("cannot find expected bookmark URL")
			}
		})
	})
	t.Run("Duplicated", func(t *testing.T) {
		t.Run("badDB", func(t *testing.T) {
			errDB := errors.New("bad DB")
			repository := &RepositoryMock{
				DuplicatedFunc: func() ([]*bookmarks.Bookmark, error) {
					return nil, errDB
				},
			}
			root := bookmarks.New(repository)
			ts := httptest.NewServer(New(root, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/duplicated")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				t.Fatal("not OK")
			}
		})
		t.Run("good", func(t *testing.T) {
			foundBookmark := &bookmarks.Bookmark{ID: 1, Title: "%FIND-TITLE%", URL: "https://%FIND-%URL.com"}
			repository := &RepositoryMock{
				DuplicatedFunc: func() ([]*bookmarks.Bookmark, error) {
					return []*bookmarks.Bookmark{
						foundBookmark,
					}, nil
				},
			}
			root := bookmarks.New(repository)
			ts := httptest.NewServer(New(root, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/duplicated")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatal("not OK")
			}
			buf := &bytes.Buffer{}
			io.Copy(buf, resp.Body)
			if !strings.Contains(buf.String(), foundBookmark.Title) {
				t.Error("cannot find expected bookmark title")
			}
			if !strings.Contains(buf.String(), foundBookmark.URL) {
				t.Error("cannot find expected bookmark URL")
			}
		})
	})
}
