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
)

type Bookmarks struct {
	repository Repository
}

func New(repository Repository) *Bookmarks {
	return &Bookmarks{
		repository: repository,
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

func (b *Bookmarks) Insert(bookmark *Bookmark, urlChecker URLChecker) error {
	if err := b.isSetup(); err != nil {
		return fmt.Errorf("cannot begin inserting bookmark: %w", err)
	}
	if urlChecker == nil {
		return errBookmarksURLCheckerNotSet
	}
	if bookmark == nil {
		return errNilBookmark
	}
	if _, err := url.Parse(bookmark.URL); err != nil {
		return &BadURLError{cause: err}
	}
	bookmark.Title, bookmark.LastStatusCheck, bookmark.LastStatusCode, bookmark.LastStatusReason = urlChecker.Check(bookmark.URL, bookmark.Title)
	if err := b.repository.Insert(bookmark); err != nil {
		return fmt.Errorf("cannot insert bookmark: %w", err)
	}
	return nil
}

func (b *Bookmarks) DeleteByID(id int64) error {
	if err := b.repository.DeleteByID(id); err != nil {
		return fmt.Errorf("cannot delete bookmark: %w", err)
	}
	return nil
}

func (b *Bookmarks) UpdateInbox(id int64, inbox string) error {
	parsedInbox, err := ParseInbox(inbox)
	if err != nil {
		return fmt.Errorf("cannot parse inbox: %w", err)
	}
	bookmark, err := b.repository.GetByID(id)
	if err != nil {
		return fmt.Errorf("cannot find bookmark: %w", err)
	}
	bookmark.Inbox = parsedInbox
	if err := b.repository.Update(bookmark); err != nil {
		return fmt.Errorf("cannot store bookmark: %w", err)
	}
	return nil
}

func (b *Bookmarks) Inbox() ([]*Bookmark, error) {
	list, err := b.repository.Inbox()
	if err != nil {
		return nil, fmt.Errorf("cannot load bookmarks inbox: %w", err)
	}
	return list, nil
}

func (b *Bookmarks) Duplicated() ([]*Bookmark, error) {
	list, err := b.repository.Duplicated()
	if err != nil {
		return nil, fmt.Errorf("cannot load duplicated bookmarks: %w", err)
	}
	return list, nil
}

func (b *Bookmarks) All() ([]*Bookmark, error) {
	list, err := b.repository.All()
	if err != nil {
		return nil, fmt.Errorf("cannot load all bookmarks: %w", err)
	}
	return list, nil
}

func (b *Bookmarks) Search(term string) ([]*Bookmark, error) {
	list, err := b.repository.Search(term)
	if err != nil {
		return nil, fmt.Errorf("cannot search bookmarks: %w", err)
	}
	return list, nil
}
