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
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"
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
	bookmark.Title, bookmark.LastStatusCheck, bookmark.LastStatusCode, bookmark.LastStatusReason = b.urlChecker.Check(bookmark.URL, bookmark.Title)
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

func (b *Bookmarks) Inbox(page int) ([]*Bookmark, error) {
	list, err := b.repository.Inbox(page)
	if err != nil {
		return nil, fmt.Errorf("cannot load bookmarks inbox: %w", err)
	}
	return list, nil
}

func (b *Bookmarks) Duplicated(page int) ([]*Bookmark, error) {
	list, err := b.repository.Duplicated(page)
	if err != nil {
		return nil, fmt.Errorf("cannot load duplicated bookmarks: %w", err)
	}
	return list, nil
}

func (b *Bookmarks) Dead(page int) ([]*Bookmark, error) {
	list, err := b.repository.Dead(page)
	if err != nil {
		return nil, fmt.Errorf("cannot load dead bookmarks: %w", err)
	}
	return list, nil
}

func (b *Bookmarks) All(page int) ([]*Bookmark, error) {
	list, err := b.repository.All(page)
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

func (b *Bookmarks) RestorePostponedLinks(ctx context.Context) error {
	err := b.repository.RestorePostponedLinks(ctx)
	if err != nil {
		return fmt.Errorf("cannot restore postponed links: %w", err)
	}
	return nil
}

func (b *Bookmarks) RefreshExpiredLinks(ctx context.Context) error {
	expiredBookmarks, err := b.repository.Expired()
	if err != nil {
		return fmt.Errorf("cannot load expired bookmarks: %w", err)
	}

	bookmarkCh := make(chan *Bookmark)
	var (
		wg        sync.WaitGroup
		muAllErrs sync.Mutex
		allErrs   error
	)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for bookmark := range bookmarkCh {
				log.Println("linkHealth:", bookmark.ID, bookmark.URL)
				bookmark.Title, bookmark.LastStatusCheck, bookmark.LastStatusCode, bookmark.LastStatusReason = b.urlChecker.Check(bookmark.URL, bookmark.Title)
				if err := b.repository.Update(bookmark); err != nil {
					muAllErrs.Lock()
					allErrs = errors.Join(allErrs, err)
					muAllErrs.Unlock()
				}
				time.Sleep(1 * time.Second)
			}
		}()
	}
	for _, bookmark := range expiredBookmarks {
		if ctx.Err() != nil {
			break
		}
		bookmarkCh <- bookmark
	}
	close(bookmarkCh)
	wg.Wait()

	return allErrs
}
