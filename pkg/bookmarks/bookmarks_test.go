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
			if err := New(tt.args.repository).DeleteByID(tt.args.id); (err != nil) != tt.wantErr {
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
