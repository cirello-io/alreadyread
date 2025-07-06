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

package bookmarks

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
)

func TestBadURLError(t *testing.T) {
	errWrap := errors.New("wrapped error")
	type fields struct {
		cause  error
		target error
	}
	tests := []struct {
		name       string
		fields     fields
		wantError  string
		wantUnwrap error
		wantIs     bool
	}{
		{"nil", fields{}, "invalid URL: <nil>", nil, false},
		{"Is/right", fields{target: &BadURLError{}}, "invalid URL: <nil>", nil, true},
		{"Is/wrong", fields{target: errors.New("")}, "invalid URL: <nil>", nil, false},
		{"Unwrap", fields{cause: errWrap}, "invalid URL: wrapped error", errWrap, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BadURLError{
				cause: tt.fields.cause,
			}
			if got := b.Error(); got != tt.wantError {
				t.Errorf("BadURLError.Error() = %v, want %v", got, tt.wantError)
			}
			if got := b.Unwrap(); !errors.Is(got, tt.wantUnwrap) {
				t.Errorf("BadURLError.Unwrap() = %v, want %v", got, tt.wantUnwrap)
			}
			if got := b.Is(tt.fields.target); got != tt.wantIs {
				t.Errorf("BadURLError.Is() = %v, want %v", got, tt.wantIs)
			}
		})
	}
}

func TestBookmarks_Insert(t *testing.T) {
	errExpectedDBError := errors.New("bad DB")
	type fields struct {
		repository Repository
		urlChecker URLChecker
	}
	type args struct {
		bookmark *Bookmark
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		expectedError error
	}{
		{"badSetup/missingRepository", fields{}, args{&Bookmark{}}, errBookmarksRepositoryNotSet},
		{"badSetup/missingURLChecker", fields{&RepositoryMock{}, nil}, args{&Bookmark{}}, errBookmarksURLCheckerNotSet},
		{"missingBookmark", fields{&RepositoryMock{}, &URLCheckerMock{}}, args{nil}, errNilBookmark},
		{"badURL", fields{&RepositoryMock{}, &URLCheckerMock{}}, args{&Bookmark{URL: "://"}}, &BadURLError{}},
		{"badDB", fields{&RepositoryMock{InsertFunc: func(*Bookmark) error { return errExpectedDBError }}, &URLCheckerMock{CheckFunc: func(_, _ string) (string, int64, int64, string) { return "", 0, 0, "" }}}, args{&Bookmark{URL: "http://example.org"}}, errExpectedDBError},
		{"good", fields{&RepositoryMock{InsertFunc: func(*Bookmark) error { return nil }}, &URLCheckerMock{CheckFunc: func(_, _ string) (string, int64, int64, string) { return "", 0, 0, "" }}}, args{&Bookmark{URL: "http://example.org"}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.fields.repository, tt.fields.urlChecker)
			if err := b.Insert(tt.args.bookmark); (err != nil) && !errors.Is(err, tt.expectedError) {
				t.Errorf("Bookmarks.Insert() error = %v, wantErr %v", err, tt.expectedError)
				return
			}
		})
	}
}

func TestBookmarks_DeleteByID(t *testing.T) {
	type args struct {
		repository Repository
		id         int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"badDelete",
			args{
				repository: &RepositoryMock{
					DeleteByIDFunc: func(int64) error {
						return errors.New("mocked error")
					},
				},
				id: 1},
			true,
		},
		{
			"goodDelete",
			args{
				repository: &RepositoryMock{
					DeleteByIDFunc: func(int64) error {
						return nil
					},
				},
				id: 1},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := New(tt.args.repository, nil).DeleteByID(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBookmarks_UpdateInbox(t *testing.T) {
	errDB := errors.New("DB error")
	foundBookmark := &Bookmark{ID: 1, Title: "title", URL: "http://url.com"}
	type fields struct {
		repository Repository
	}
	type args struct {
		id    int64
		inbox string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"badInbox", fields{}, args{0, "bad"}, true},
		{"badDB/Get", fields{repository: &RepositoryMock{GetByIDFunc: func(int64) (*Bookmark, error) { return nil, errDB }}}, args{0, "new"}, true},
		{"badDB/Update", fields{repository: &RepositoryMock{GetByIDFunc: func(int64) (*Bookmark, error) { return foundBookmark, nil }, UpdateFunc: func(*Bookmark) error { return errDB }}}, args{foundBookmark.ID, "new"}, true},
		{
			"done",
			fields{
				repository: &RepositoryMock{
					GetByIDFunc: func(int64) (*Bookmark, error) {
						return foundBookmark, nil
					},
					UpdateFunc: func(bookmark *Bookmark) error {
						if bookmark != foundBookmark {
							t.Error("unexpected bookmark used in update")
						}
						return nil
					},
				},
			},
			args{foundBookmark.ID, "new"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bookmarks{
				repository: tt.fields.repository,
			}
			if err := b.UpdateInbox(tt.args.id, tt.args.inbox); (err != nil) != tt.wantErr {
				t.Errorf("Bookmarks.UpdateInbox() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBookmarks_Inbox(t *testing.T) {
	errDB := errors.New("DB error")
	foundBookmark := &Bookmark{ID: 1, Title: "title", URL: "http://url.com"}
	type fields struct {
		repository Repository
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*Bookmark
		wantErr bool
	}{
		{"badDB", fields{repository: &RepositoryMock{InboxFunc: func() ([]*Bookmark, error) { return nil, errDB }}}, nil, true},
		{"nilResult", fields{repository: &RepositoryMock{InboxFunc: func() ([]*Bookmark, error) { return nil, nil }}}, nil, false},
		{"emptyResult", fields{repository: &RepositoryMock{InboxFunc: func() ([]*Bookmark, error) { return []*Bookmark{}, nil }}}, []*Bookmark{}, false},
		{"good", fields{repository: &RepositoryMock{InboxFunc: func() ([]*Bookmark, error) { return []*Bookmark{foundBookmark}, nil }}}, []*Bookmark{foundBookmark}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bookmarks{
				repository: tt.fields.repository,
			}
			got, err := b.Inbox()
			if (err != nil) != tt.wantErr {
				t.Errorf("Bookmarks.Inbox() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bookmarks.Inbox() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBookmarks_Duplicated(t *testing.T) {
	errDB := errors.New("DB error")
	foundBookmark := &Bookmark{ID: 1, Title: "title", URL: "http://url.com"}
	type fields struct {
		repository Repository
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*Bookmark
		wantErr bool
	}{
		{"badDB", fields{repository: &RepositoryMock{DuplicatedFunc: func() ([]*Bookmark, error) { return nil, errDB }}}, nil, true},
		{"nilResult", fields{repository: &RepositoryMock{DuplicatedFunc: func() ([]*Bookmark, error) { return nil, nil }}}, nil, false},
		{"emptyResult", fields{repository: &RepositoryMock{DuplicatedFunc: func() ([]*Bookmark, error) { return []*Bookmark{}, nil }}}, []*Bookmark{}, false},
		{"good", fields{repository: &RepositoryMock{DuplicatedFunc: func() ([]*Bookmark, error) { return []*Bookmark{foundBookmark}, nil }}}, []*Bookmark{foundBookmark}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bookmarks{
				repository: tt.fields.repository,
			}
			got, err := b.Duplicated()
			if (err != nil) != tt.wantErr {
				t.Errorf("Bookmarks.Duplicated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bookmarks.Duplicated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBookmarks_Dead(t *testing.T) {
	errDB := errors.New("DB error")
	foundBookmark := &Bookmark{ID: 1, Title: "title", URL: "http://url.com", LastStatusCode: http.StatusInternalServerError}
	type fields struct {
		repository Repository
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*Bookmark
		wantErr bool
	}{
		{"badDB", fields{repository: &RepositoryMock{DeadFunc: func() ([]*Bookmark, error) { return nil, errDB }}}, nil, true},
		{"nilResult", fields{repository: &RepositoryMock{DeadFunc: func() ([]*Bookmark, error) { return nil, nil }}}, nil, false},
		{"emptyResult", fields{repository: &RepositoryMock{DeadFunc: func() ([]*Bookmark, error) { return []*Bookmark{}, nil }}}, []*Bookmark{}, false},
		{"good", fields{repository: &RepositoryMock{DeadFunc: func() ([]*Bookmark, error) { return []*Bookmark{foundBookmark}, nil }}}, []*Bookmark{foundBookmark}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bookmarks{
				repository: tt.fields.repository,
			}
			got, err := b.Dead()
			if (err != nil) != tt.wantErr {
				t.Errorf("Bookmarks.Dead() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bookmarks.Dead() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBookmarks_All(t *testing.T) {
	errDB := errors.New("DB error")
	foundBookmark := &Bookmark{ID: 1, Title: "title", URL: "http://url.com"}
	type fields struct {
		repository Repository
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*Bookmark
		wantErr bool
	}{
		{"badDB", fields{repository: &RepositoryMock{AllFunc: func(int) ([]*Bookmark, error) { return nil, errDB }}}, nil, true},
		{"nilResult", fields{repository: &RepositoryMock{AllFunc: func(int) ([]*Bookmark, error) { return nil, nil }}}, nil, false},
		{"emptyResult", fields{repository: &RepositoryMock{AllFunc: func(int) ([]*Bookmark, error) { return []*Bookmark{}, nil }}}, []*Bookmark{}, false},
		{"good", fields{repository: &RepositoryMock{AllFunc: func(int) ([]*Bookmark, error) { return []*Bookmark{foundBookmark}, nil }}}, []*Bookmark{foundBookmark}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bookmarks{
				repository: tt.fields.repository,
			}
			got, err := b.All(0)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bookmarks.All() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bookmarks.All() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBookmarks_Search(t *testing.T) {
	errDB := errors.New("DB error")
	foundBookmark := &Bookmark{ID: 1, Title: "title", URL: "http://url.com"}
	type fields struct {
		repository Repository
	}
	type args struct {
		term string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Bookmark
		wantErr bool
	}{
		{"badDB", fields{repository: &RepositoryMock{SearchFunc: func(string) ([]*Bookmark, error) { return nil, errDB }}}, args{}, nil, true},
		{"nilResult", fields{repository: &RepositoryMock{SearchFunc: func(string) ([]*Bookmark, error) { return nil, nil }}}, args{}, nil, false},
		{"emptyResult", fields{repository: &RepositoryMock{SearchFunc: func(string) ([]*Bookmark, error) { return []*Bookmark{}, nil }}}, args{}, []*Bookmark{}, false},
		{"good", fields{repository: &RepositoryMock{SearchFunc: func(string) ([]*Bookmark, error) { return []*Bookmark{foundBookmark}, nil }}}, args{}, []*Bookmark{foundBookmark}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bookmarks{
				repository: tt.fields.repository,
			}
			got, err := b.Search(tt.args.term)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bookmarks.Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bookmarks.Search() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBookmarks_RestorePostponedLinks(t *testing.T) {
	errDB := errors.New("DB error")
	type fields struct {
		repository Repository
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"badDB", fields{repository: &RepositoryMock{RestorePostponedLinksFunc: func(context.Context) error { return errDB }}}, true},
		{"good", fields{repository: &RepositoryMock{RestorePostponedLinksFunc: func(context.Context) error { return nil }}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bookmarks{
				repository: tt.fields.repository,
			}
			err := b.RestorePostponedLinks(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Bookmarks.Dead() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestBookmarks_RefreshExpiredLinks(t *testing.T) {
	t.Run("badDB/expiration", func(t *testing.T) {
		errDB := errors.New("bad DB")
		repository := &RepositoryMock{
			ExpiredFunc: func() ([]*Bookmark, error) {
				return nil, errDB
			},
		}
		urlchecker := &URLCheckerMock{}
		b := New(repository, urlchecker)
		err := b.RefreshExpiredLinks(context.TODO())
		if !errors.Is(err, errDB) {
			t.Fatal("unexpected error:", err)
		}
	})
	t.Run("badDB/update", func(t *testing.T) {
		errDB := errors.New("bad DB")
		foundBookmarks := []*Bookmark{{ID: 1, URL: "https://example.com"}}
		repository := &RepositoryMock{
			ExpiredFunc: func() ([]*Bookmark, error) {
				return foundBookmarks, nil
			},
			UpdateFunc: func(*Bookmark) error {
				return errDB
			},
		}
		const expectedTitle = "title"
		urlchecker := &URLCheckerMock{
			CheckFunc: func(_, _ string) (string, int64, int64, string) {
				return expectedTitle, 0, 0, ""
			},
		}
		b := New(repository, urlchecker)
		err := b.RefreshExpiredLinks(context.TODO())
		if !errors.Is(err, errDB) {
			t.Fatal("unexpected error:", err)
		}
	})
	t.Run("canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		foundBookmarks := []*Bookmark{{ID: 1, URL: "https://example.com"}}
		repository := &RepositoryMock{
			ExpiredFunc: func() ([]*Bookmark, error) {
				cancel()
				return foundBookmarks, nil
			},
			UpdateFunc: func(*Bookmark) error {
				t.Fatal("unexpected update")
				return nil
			},
		}
		const expectedTitle = "title"
		urlchecker := &URLCheckerMock{
			CheckFunc: func(_, _ string) (string, int64, int64, string) {
				return expectedTitle, 0, 0, ""
			},
		}
		b := New(repository, urlchecker)
		err := b.RefreshExpiredLinks(ctx)
		if err != nil {
			t.Fatal("unexpected error:", err)
		}
	})
	t.Run("good", func(t *testing.T) {
		foundBookmarks := []*Bookmark{{ID: 1, URL: "https://example.com"}}
		repository := &RepositoryMock{
			ExpiredFunc: func() ([]*Bookmark, error) {
				return foundBookmarks, nil
			},
			UpdateFunc: func(*Bookmark) error {
				return nil
			},
		}
		const expectedTitle = "title"
		urlchecker := &URLCheckerMock{
			CheckFunc: func(_, _ string) (string, int64, int64, string) {
				return expectedTitle, 0, 0, ""
			},
		}
		b := New(repository, urlchecker)
		err := b.RefreshExpiredLinks(context.TODO())
		if err != nil {
			t.Fatal("unexpected error:", err)
		}
	})
}
