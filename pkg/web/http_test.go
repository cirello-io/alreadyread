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
	"net/url"
	"strings"
	"testing"

	"cirello.io/alreadyread/frontend"
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
	t.Run("post", func(t *testing.T) {
		t.Run("emptyTitle", func(t *testing.T) {
			errDB := errors.New("bad DB")
			repository := &RepositoryMock{
				InboxFunc: func() ([]*bookmarks.Bookmark, error) {
					return nil, errDB
				},
			}
			root := bookmarks.New(repository, nil)

			ts := httptest.NewServer(New(root, &URLCheckerMock{
				TitleFunc: func(string) string {
					return "example-title"
				},
			}, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Post(ts.URL+"/post?loadTitle=true&url="+url.QueryEscape("https://example.com"), "", nil)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatal("not OK")
			}

			buf := &bytes.Buffer{}
			_, _ = io.Copy(buf, resp.Body)
			if !strings.Contains(buf.String(), "example-title") {
				t.Error("cannot find expected bookmark title")
			}
			if !strings.Contains(buf.String(), "https://example.com") {
				t.Error("cannot find expected bookmark URL")
			}
		})
	})
	t.Run("inbox", func(t *testing.T) {
		t.Run("badDB", func(t *testing.T) {
			errDB := errors.New("bad DB")
			repository := &RepositoryMock{
				InboxFunc: func() ([]*bookmarks.Bookmark, error) {
					return nil, errDB
				},
			}
			root := bookmarks.New(repository, nil)
			ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/inbox")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusInternalServerError {
				t.Fatal("not StatusInternalServerError:", resp.StatusCode)
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
			root := bookmarks.New(repository, nil)
			ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/inbox")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatal("not OK:", resp.StatusCode)
			}
			buf := &bytes.Buffer{}
			_, _ = io.Copy(buf, resp.Body)
			if !strings.Contains(buf.String(), foundBookmark.Title) {
				t.Error("cannot find expected bookmark title")
			}
			if !strings.Contains(buf.String(), foundBookmark.URL) {
				t.Error("cannot find expected bookmark URL")
			}
		})
	})
	t.Run("duplicated", func(t *testing.T) {
		t.Run("badDB", func(t *testing.T) {
			errDB := errors.New("bad DB")
			repository := &RepositoryMock{
				DuplicatedFunc: func() ([]*bookmarks.Bookmark, error) {
					return nil, errDB
				},
			}
			root := bookmarks.New(repository, nil)
			ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/duplicated")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusInternalServerError {
				t.Fatal("not StatusInternalServerError:", resp.StatusCode)
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
			root := bookmarks.New(repository, nil)
			ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/duplicated")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatal("not OK:", resp.StatusCode)
			}
			buf := &bytes.Buffer{}
			_, _ = io.Copy(buf, resp.Body)
			if !strings.Contains(buf.String(), foundBookmark.Title) {
				t.Error("cannot find expected bookmark title")
			}
			if !strings.Contains(buf.String(), foundBookmark.URL) {
				t.Error("cannot find expected bookmark URL")
			}
		})
	})
	t.Run("dead", func(t *testing.T) {
		t.Run("badDB", func(t *testing.T) {
			errDB := errors.New("bad DB")
			repository := &RepositoryMock{
				DeadFunc: func() ([]*bookmarks.Bookmark, error) {
					return nil, errDB
				},
			}
			root := bookmarks.New(repository, nil)
			ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/dead")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusInternalServerError {
				t.Fatal("not StatusInternalServerError:", resp.StatusCode)
			}
		})
		t.Run("good", func(t *testing.T) {
			foundBookmark := &bookmarks.Bookmark{ID: 1, Title: "%FIND-TITLE%", URL: "https://%FIND-%URL.com"}
			repository := &RepositoryMock{
				DeadFunc: func() ([]*bookmarks.Bookmark, error) {
					return []*bookmarks.Bookmark{
						foundBookmark,
					}, nil
				},
			}
			root := bookmarks.New(repository, nil)
			ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/dead")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatal("not OK:", resp.StatusCode)
			}
			buf := &bytes.Buffer{}
			_, _ = io.Copy(buf, resp.Body)
			if !strings.Contains(buf.String(), foundBookmark.Title) {
				t.Error("cannot find expected bookmark title")
			}
			if !strings.Contains(buf.String(), foundBookmark.URL) {
				t.Error("cannot find expected bookmark URL")
			}
		})
	})
	t.Run("all", func(t *testing.T) {
		t.Run("badDB", func(t *testing.T) {
			errDB := errors.New("bad DB")
			repository := &RepositoryMock{
				AllFunc: func() ([]*bookmarks.Bookmark, error) {
					return nil, errDB
				},
			}
			root := bookmarks.New(repository, nil)
			ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/all")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusInternalServerError {
				t.Fatal("not StatusInternalServerError:", resp.StatusCode)
			}
		})
		t.Run("good", func(t *testing.T) {
			foundBookmark := &bookmarks.Bookmark{ID: 1, Title: "%FIND-TITLE%", URL: "https://%FIND-%URL.com"}
			repository := &RepositoryMock{
				AllFunc: func() ([]*bookmarks.Bookmark, error) {
					return []*bookmarks.Bookmark{
						foundBookmark,
					}, nil
				},
			}
			root := bookmarks.New(repository, nil)
			ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/all")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatal("not OK:", resp.StatusCode)
			}
			buf := &bytes.Buffer{}
			_, _ = io.Copy(buf, resp.Body)
			if !strings.Contains(buf.String(), foundBookmark.Title) {
				t.Error("cannot find expected bookmark title")
			}
			if !strings.Contains(buf.String(), foundBookmark.URL) {
				t.Error("cannot find expected bookmark URL")
			}
		})
	})
	t.Run("search", func(t *testing.T) {
		t.Run("badDB", func(t *testing.T) {
			errDB := errors.New("bad DB")
			repository := &RepositoryMock{
				SearchFunc: func(string) ([]*bookmarks.Bookmark, error) {
					return nil, errDB
				},
			}
			root := bookmarks.New(repository, nil)
			ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/search?term=banana")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusInternalServerError {
				t.Fatal("not StatusInternalServerError:", resp.StatusCode)
			}
		})
		t.Run("good", func(t *testing.T) {
			foundBookmark := &bookmarks.Bookmark{ID: 1, Title: "%FIND-TITLE%", URL: "https://%FIND-%URL.com"}
			const expectedTerm = "banana"
			repository := &RepositoryMock{
				SearchFunc: func(term string) ([]*bookmarks.Bookmark, error) {
					if term != expectedTerm {
						t.Error("unexpected term found:", term)
					}
					return []*bookmarks.Bookmark{
						foundBookmark,
					}, nil
				},
			}
			root := bookmarks.New(repository, nil)
			ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/search?term=" + expectedTerm)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatal("not OK:", resp.StatusCode)
			}
			buf := &bytes.Buffer{}
			_, _ = io.Copy(buf, resp.Body)
			if !strings.Contains(buf.String(), foundBookmark.Title) {
				t.Error("cannot find expected bookmark title")
			}
			if !strings.Contains(buf.String(), foundBookmark.URL) {
				t.Error("cannot find expected bookmark URL")
			}
		})
	})
	t.Run("operations", func(t *testing.T) {
		t.Run("badID", func(t *testing.T) {
			root := bookmarks.New(nil, nil)
			ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/bookmarks/badID/")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatal("not StatusBadRequest:", resp.StatusCode)
			}
		})
		t.Run("badMethod", func(t *testing.T) {
			root := bookmarks.New(nil, nil)
			ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
			defer ts.Close()
			resp, err := ts.Client().Get(ts.URL + "/bookmarks/1/")
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Fatal("not StatusMethodNotAllowed:", resp.StatusCode)
			}
		})
		t.Run("methodDelete", func(t *testing.T) {
			t.Run("badDB", func(t *testing.T) {
				errDB := errors.New("bad DB")
				repository := &RepositoryMock{
					DeleteByIDFunc: func(int64) error {
						return errDB
					},
				}
				root := bookmarks.New(repository, nil)
				ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
				defer ts.Close()
				req, err := http.NewRequest(http.MethodDelete, ts.URL+"/bookmarks/1/", nil)
				if err != nil {
					t.Fatal(err)
				}
				resp, err := ts.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusInternalServerError {
					t.Fatal("not StatusInternalServerError:", resp.StatusCode)
				}
			})
			t.Run("good", func(t *testing.T) {
				repository := &RepositoryMock{
					DeleteByIDFunc: func(int64) error {
						return nil
					},
				}
				root := bookmarks.New(repository, nil)
				ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
				defer ts.Close()
				req, err := http.NewRequest(http.MethodDelete, ts.URL+"/bookmarks/1/", nil)
				if err != nil {
					t.Fatal(err)
				}
				resp, err := ts.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Fatal("not StatusOK:", resp.StatusCode)
				}
			})
		})
		t.Run("methodPatch", func(t *testing.T) {
			t.Run("badInbox", func(t *testing.T) {
				root := bookmarks.New(nil, nil)
				ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
				defer ts.Close()
				req, err := http.NewRequest(http.MethodPatch, ts.URL+"/bookmarks/1/?inbox=banana", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				resp, err := ts.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusInternalServerError {
					t.Fatal("not StatusInternalServerError:", resp.StatusCode)
				}
			})
			t.Run("badDB/GetByID", func(t *testing.T) {
				errDB := errors.New("bad DB")
				repository := &RepositoryMock{
					GetByIDFunc: func(int64) (*bookmarks.Bookmark, error) { return nil, errDB },
				}
				root := bookmarks.New(repository, nil)
				ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
				defer ts.Close()
				req, err := http.NewRequest(http.MethodPatch, ts.URL+"/bookmarks/1/?inbox=read", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				resp, err := ts.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusInternalServerError {
					t.Fatal("not StatusInternalServerError:", resp.StatusCode)
				}
			})
			t.Run("badDB/Update", func(t *testing.T) {
				errDB := errors.New("bad DB")
				foundBookmark := &bookmarks.Bookmark{
					URL:   "https://example.com",
					Title: "title",
				}
				repository := &RepositoryMock{
					GetByIDFunc: func(int64) (*bookmarks.Bookmark, error) { return foundBookmark, nil },
					UpdateFunc:  func(*bookmarks.Bookmark) error { return errDB },
				}
				root := bookmarks.New(repository, nil)
				ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
				defer ts.Close()
				req, err := http.NewRequest(http.MethodPatch, ts.URL+"/bookmarks/1/?inbox=read", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				resp, err := ts.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusInternalServerError {
					t.Fatal("not StatusInternalServerError:", resp.StatusCode)
				}
			})
			t.Run("good", func(t *testing.T) {
				foundBookmark := &bookmarks.Bookmark{
					URL:   "https://example.com",
					Title: "title",
				}
				repository := &RepositoryMock{
					GetByIDFunc: func(int64) (*bookmarks.Bookmark, error) { return foundBookmark, nil },
					UpdateFunc: func(*bookmarks.Bookmark) error {
						return nil
					},
				}
				root := bookmarks.New(repository, nil)
				ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
				defer ts.Close()
				req, err := http.NewRequest(http.MethodPatch, ts.URL+"/bookmarks/1/?inbox=postponed", nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				resp, err := ts.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Fatal("not StatusOK:", resp.StatusCode)
				}
				if foundBookmark.Inbox != bookmarks.Postponed {
					t.Fatal("did not update")
				}
			})
		})
		t.Run("methodPost", func(t *testing.T) {
			t.Run("emptyBookmark", func(t *testing.T) {
				root := bookmarks.New(nil, nil)
				ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
				defer ts.Close()
				form := url.Values{
					"title": {},
					"url":   {},
				}
				req, err := http.NewRequest(http.MethodPost, ts.URL+"/bookmarks/1/", strings.NewReader(form.Encode()))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				resp, err := ts.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusBadRequest {
					t.Fatal("not StatusBadRequest:", resp.StatusCode)
				}
			})
			t.Run("badDB/Insert", func(t *testing.T) {
				errDB := errors.New("bad DB")
				repository := &RepositoryMock{
					InsertFunc: func(*bookmarks.Bookmark) error {
						return errDB
					},
				}
				urlChecker := &URLCheckerMock{
					CheckFunc: func(_, _ string) (string, int64, int64, string) {
						return "title", 0, 0, ""
					},
				}
				root := bookmarks.New(repository, urlChecker)
				ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
				defer ts.Close()
				form := url.Values{
					"title": {"title"},
					"url":   {"https://example.com"},
				}
				req, err := http.NewRequest(http.MethodPost, ts.URL+"/bookmarks/1/", strings.NewReader(form.Encode()))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				resp, err := ts.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusInternalServerError {
					t.Fatal("not StatusInternalServerError:", resp.StatusCode)
				}
			})
			t.Run("good", func(t *testing.T) {
				repository := &RepositoryMock{
					InsertFunc: func(*bookmarks.Bookmark) error {
						return nil
					},
					InboxFunc: func() ([]*bookmarks.Bookmark, error) {
						return []*bookmarks.Bookmark{
							{ID: 1, Title: "title", URL: "https://example.com"},
						}, nil
					},
				}
				urlChecker := &URLCheckerMock{
					CheckFunc: func(_, _ string) (string, int64, int64, string) {
						return "title", 0, 0, ""
					},
				}
				root := bookmarks.New(repository, urlChecker)
				ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
				defer ts.Close()
				form := url.Values{
					"title": {"title"},
					"url":   {"https://example.com"},
				}
				req, err := http.NewRequest(http.MethodPost, ts.URL+"/bookmarks/1/", strings.NewReader(form.Encode()))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				resp, err := ts.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					t.Fatal("not StatusOK:", resp.StatusCode)
				}
			})
		})
	})
	t.Run("index", func(t *testing.T) {
		root := bookmarks.New(nil, nil)
		ts := httptest.NewServer(New(root, nil, []string{"localhost"}))
		defer ts.Close()
		resp, err := ts.Client().Get(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatal("not OK:", resp.StatusCode)
		}
		respBuf := &bytes.Buffer{}
		_, _ = io.Copy(respBuf, resp.Body)
		tplBuf := &bytes.Buffer{}
		frontend.RenderIndex(tplBuf, "/inbox", "")
		if tplBuf.String() != respBuf.String() {
			t.Fatal("index page not rendering correctly")
		}
	})
}
