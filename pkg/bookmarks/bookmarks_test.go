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

func TestDeleteByID(t *testing.T) {
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
			if err := DeleteByID(tt.args.repository, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteByID() error = %v, wantErr %v", err, tt.wantErr)
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
		want          *Bookmark
		expectedError error
	}{
		{"badSetup/missingRepository", fields{}, args{&Bookmark{}}, nil, errBookmarksRepositoryNotSet},
		{"badSetup/missingURLChecker", fields{&RepositoryMock{}, nil}, args{&Bookmark{}}, nil, errBookmarksURLCheckerNotSet},
		{"missingBookmark", fields{&RepositoryMock{}, &URLCheckerMock{}}, args{nil}, nil, errNilBookmark},
		{"badURL", fields{&RepositoryMock{}, &URLCheckerMock{}}, args{&Bookmark{URL: "://"}}, nil, &BadURLError{}},
		{"badDB", fields{&RepositoryMock{InsertFunc: func(bookmark *Bookmark) (*Bookmark, error) { return nil, errExpectedDBError }}, &URLCheckerMock{CheckFunc: func(bookmark *Bookmark) *Bookmark { return bookmark }}}, args{&Bookmark{URL: "http://example.org"}}, nil, errExpectedDBError},
		{"good", fields{&RepositoryMock{InsertFunc: func(bookmark *Bookmark) (*Bookmark, error) { return bookmark, nil }}, &URLCheckerMock{CheckFunc: func(bookmark *Bookmark) *Bookmark { bookmark.Title = "Title"; return bookmark }}}, args{&Bookmark{URL: "http://example.org"}}, &Bookmark{Title: "Title", URL: "http://example.org"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bookmarks{
				repository: tt.fields.repository,
				urlChecker: tt.fields.urlChecker,
			}
			got, err := b.Insert(tt.args.bookmark)
			if (err != nil) && !errors.Is(err, tt.expectedError) {
				t.Errorf("Bookmarks.Insert() error = %v, wantErr %v", err, tt.expectedError)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bookmarks.Insert() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
			if got := b.Unwrap(); got != tt.wantUnwrap {
				t.Errorf("BadURLError.Unwrap() = %v, want %v", got, tt.wantUnwrap)
			}
			if got := b.Is(tt.fields.target); got != tt.wantIs {
				t.Errorf("BadURLError.Is() = %v, want %v", got, tt.wantIs)
			}
		})
	}
}
