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
			"badDB",
			args{
				repository: &RepositoryMock{
					GetByIDFunc: func(id int64) (*Bookmark, error) {
						return nil, errors.New("mocked error")
					},
				},
				id: 0},
			true,
		},
		{
			"badDelete",
			args{
				repository: &RepositoryMock{
					GetByIDFunc: func(id int64) (*Bookmark, error) {
						return &Bookmark{ID: id}, nil
					},
					DeleteFunc: func(bookmark *Bookmark) error {
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
					GetByIDFunc: func(id int64) (*Bookmark, error) {
						return &Bookmark{ID: id}, nil
					},
					DeleteFunc: func(bookmark *Bookmark) error {
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
