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
	"errors"
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
		{"badDB", fields{&RepositoryMock{InsertFunc: func(bookmark *Bookmark) error { return errExpectedDBError }}, &URLCheckerMock{CheckFunc: func(url, originalTitle string) (string, int64, int64, string) { return "", 0, 0, "" }}}, args{&Bookmark{URL: "http://example.org"}}, errExpectedDBError},
		{"good", fields{&RepositoryMock{InsertFunc: func(bookmark *Bookmark) error { return nil }}, &URLCheckerMock{CheckFunc: func(url, originalTitle string) (string, int64, int64, string) { return "", 0, 0, "" }}}, args{&Bookmark{URL: "http://example.org"}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := New(tt.fields.repository)
			if err := b.Insert(tt.args.bookmark, tt.fields.urlChecker); (err != nil) && !errors.Is(err, tt.expectedError) {
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
					DeleteByIDFunc: func(id int64) error {
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
					DeleteByIDFunc: func(id int64) error {
						return nil
					},
				},
				id: 1},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := New(tt.args.repository).DeleteByID(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteByID() error = %v, wantErr %v", err, tt.wantErr)
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
		{"badDB", fields{repository: &RepositoryMock{AllFunc: func() ([]*Bookmark, error) { return nil, errDB }}}, nil, true},
		{"nilResult", fields{repository: &RepositoryMock{AllFunc: func() ([]*Bookmark, error) { return nil, nil }}}, nil, false},
		{"emptyResult", fields{repository: &RepositoryMock{AllFunc: func() ([]*Bookmark, error) { return []*Bookmark{}, nil }}}, []*Bookmark{}, false},
		{"good", fields{repository: &RepositoryMock{AllFunc: func() ([]*Bookmark, error) { return []*Bookmark{foundBookmark}, nil }}}, []*Bookmark{foundBookmark}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bookmarks{
				repository: tt.fields.repository,
			}
			got, err := b.All()
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
		{"badDB", fields{repository: &RepositoryMock{SearchFunc: func(term string) ([]*Bookmark, error) { return nil, errDB }}}, args{}, nil, true},
		{"nilResult", fields{repository: &RepositoryMock{SearchFunc: func(term string) ([]*Bookmark, error) { return nil, nil }}}, args{}, nil, false},
		{"emptyResult", fields{repository: &RepositoryMock{SearchFunc: func(term string) ([]*Bookmark, error) { return []*Bookmark{}, nil }}}, args{}, []*Bookmark{}, false},
		{"good", fields{repository: &RepositoryMock{SearchFunc: func(term string) ([]*Bookmark, error) { return []*Bookmark{foundBookmark}, nil }}}, args{}, []*Bookmark{foundBookmark}, false},
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
