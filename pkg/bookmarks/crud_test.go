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
			if err := DeleteByID(tt.args.repository, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteByID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBookmark_Insert(t *testing.T) {
	type args struct {
		repository Repository
		b          *Bookmark
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"badURL",
			args{
				repository: &RepositoryMock{},
				b:          &Bookmark{URL: "://"},
			},
			true,
		},
		{
			"badDB",
			args{
				repository: &RepositoryMock{
					InsertFunc: func(bookmark *Bookmark) (*Bookmark, error) {
						return nil, errors.New("mocked error")
					},
				},
				b: &Bookmark{Title: "Example.com", URL: "http://example.com"},
			},
			true,
		},
		{
			"good",
			args{
				repository: &RepositoryMock{
					InsertFunc: func(bookmark *Bookmark) (*Bookmark, error) {
						bookmark.ID = 1
						return bookmark, nil
					},
				},
				b: &Bookmark{Title: "Example.com", URL: "http://example.com"},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookmark := tt.args.b
			err := bookmark.Insert(tt.args.repository)
			if (err != nil) != tt.wantErr {
				t.Errorf("Insert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
