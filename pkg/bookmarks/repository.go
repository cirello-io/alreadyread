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

//go:generate moq -out repository_mocks_test.go . Repository
type Repository interface {
	// All returns all known bookmarks.
	All() ([]*Bookmark, error)

	// Bootstrap creates table if missing.
	Bootstrap() error

	// DeleteByID excludes the bookmark from the repository.
	DeleteByID(id int64) error

	// Expired return all valid but expired bookmarks.
	Expired() ([]*Bookmark, error)

	// GetByID loads one bookmark.
	GetByID(id int64) (*Bookmark, error)

	// Insert one bookmark.
	Insert(*Bookmark) (*Bookmark, error)

	// Invalid return all invalid bookmarks.
	Invalid() ([]*Bookmark, error)

	// Update one bookmark.
	Update(*Bookmark) error
}