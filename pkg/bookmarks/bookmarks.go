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
	"fmt"
	"net/url"
	"slices"
)

type Bookmarks struct {
	repository Repository
	urlChecker URLChecker
}

func New(repository Repository, urlChecker URLChecker) *Bookmarks {
	return &Bookmarks{
		repository: repository,
		urlChecker: urlChecker,
	}
}

var (
	errBookmarksRepositoryNotSet = fmt.Errorf("repository is not set")
	errBookmarksURLCheckerNotSet = fmt.Errorf("url checker is not set")
)

func (b *Bookmarks) isSetup() error {
	if b.repository == nil {
		return errBookmarksRepositoryNotSet
	}
	if b.urlChecker == nil {
		return errBookmarksURLCheckerNotSet
	}
	return nil
}

var (
	errNilBookmark = fmt.Errorf("cannot insert nil bookmark")
)

type BadURLError struct {
	cause error
}

func (b BadURLError) Error() string {
	return fmt.Sprintf("invalid URL: %v", b.cause)
}

func (b BadURLError) Unwrap() error {
	return b.cause
}

func (b BadURLError) Is(target error) bool {
	errBadURL := &BadURLError{}
	return errors.As(target, &errBadURL)
}

func (b *Bookmarks) Insert(bookmark *Bookmark) error {
	if err := b.isSetup(); err != nil {
		return fmt.Errorf("cannot begin inserting bookmark: %w", err)
	}
	if bookmark == nil {
		return errNilBookmark
	}
	if _, err := url.Parse(bookmark.URL); err != nil {
		return &BadURLError{cause: err}
	}
	b.urlChecker.Check(bookmark)
	if err := b.repository.Insert(bookmark); err != nil {
		return fmt.Errorf("cannot insert bookmark: %w", err)
	}
	return nil
}

func DeleteByID(repository Repository, id int64) error {
	if err := repository.DeleteByID(id); err != nil {
		return fmt.Errorf("cannot delete bookmark: %w", err)
	}
	return nil
}

func UpdateInbox(repository Repository, id int64, inbox string) error {
	parsedInbox, err := ParseInbox(inbox)
	if err != nil {
		return fmt.Errorf("cannot parse inbox: %w", err)
	}
	b, err := repository.GetByID(id)
	if err != nil {
		return fmt.Errorf("cannot find bookmark: %w", err)
	}
	b.Inbox = parsedInbox
	if err := repository.Update(b); err != nil {
		return fmt.Errorf("cannot store bookmark: %w", err)
	}
	return nil
}

func List(repository Repository, filter string) ([]*Bookmark, error) {
	// TODO: use specifications and database filtering
	list, err := repository.All()
	if err != nil {
		return nil, fmt.Errorf("cannot load all bookmarks: %w", err)
	}
	switch filter {
	case "new":
		list = slices.DeleteFunc(list, func(bookmark *Bookmark) bool {
			return bookmark.Inbox != NewLink
		})
	case "duplicated":
		duplicates := map[string][]*Bookmark{}
		for _, bookmark := range list {
			duplicates[bookmark.URL] = append(duplicates[bookmark.URL], bookmark)
		}
		list = nil
		for _, duplicate := range duplicates {
			if len(duplicate) > 1 {
				list = append(list, duplicate...)
			}
		}
		slices.SortFunc(list, func(a, b *Bookmark) int {
			return b.CreatedAt.Compare(a.CreatedAt)
		})
	}

	return list, nil
}
