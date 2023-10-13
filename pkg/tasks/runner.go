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

package tasks // import "cirello.io/alreadyread/pkg/tasks"

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"cirello.io/alreadyread/pkg/bookmarks"
	"cirello.io/alreadyread/pkg/bookmarks/sqliterepo"
	"golang.org/x/sync/singleflight"
)

// Task represents one periodic task executed by the runner. Key must be unique,
// as it is used as key to lock.
type Task struct {
	Name      string
	Exec      func(db *sql.DB) error
	Frequency time.Duration
}

var execGroup singleflight.Group
var tasks = []Task{
	{"check link health", LinkHealth, 6 * time.Hour},
	{"vacuum", Vacuum, 12 * time.Hour},
	{"restore postponed links", RestorePostponedLinks, 6 * time.Hour},
}

// Run executes background maintenance tasks.
func Run(db *sql.DB) {
	run(context.Background(), db, tasks)
}

func run(ctx context.Context, db *sql.DB, tasks []Task) {
	for _, t := range tasks {
		t := t
		go func() {
			log.Println("tasks: scheduled", t.Name)
			for {
				go func() {
					_, err, _ := execGroup.Do(t.Name, func() (interface{}, error) {
						log.Println("tasks:", t.Name, "running")
						defer log.Println("tasks:", t.Name, "done")
						err := t.Exec(db)
						return nil, err
					})
					if err != nil {
						log.Println(t.Name, " failed:", err)
					}
				}()
				select {
				case <-time.After(t.Frequency):
				case <-ctx.Done():
				}
			}
		}()
	}
}

// LinkHealth checks if the expired links are still valid.
func LinkHealth(db *sql.DB) (err error) {
	defer recoverPanic(&err)
	repository := sqliterepo.New(db)
	expiredBookmarks, err := repository.Expired()
	if err != nil {
		return fmt.Errorf("cannot load expired bookmarks: %w", err)
	}

	bookmarkCh := make(chan *bookmarks.Bookmark)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			urlChecker := bookmarks.NewURLChecker()
			for bookmark := range bookmarkCh {
				log.Println("linkHealth:", bookmark.ID, bookmark.URL)
				bookmark = urlChecker.Check(bookmark)
				if err := repository.Update(bookmark); err != nil {
					log.Println(err, "cannot update link during link health check - status OK")
				}
			}
		}()
	}
	for _, bookmark := range expiredBookmarks {
		bookmarkCh <- bookmark
	}
	close(bookmarkCh)
	wg.Wait()

	return nil
}

// Vacuum executes a SQLite3 vacuum clean up.
func Vacuum(db *sql.DB) (err error) {
	defer recoverPanic(&err)
	_, err = db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("cannot run vacuum: %w", err)
	}
	return nil
}

// RestorePostponedLinks revamp rescheduled links in the inbox.
func RestorePostponedLinks(db *sql.DB) (err error) {
	defer recoverPanic(&err)
	_, err = db.Exec("UPDATE bookmarks SET inbox = 1 WHERE inbox = 2")
	if err != nil {
		return fmt.Errorf("cannot run restore rescheduled links: %w", err)
	}
	return nil
}

func recoverPanic(outErr *error) {
	if r := recover(); r != nil {
		*outErr = fmt.Errorf("recovered panic: %v", r)
	}
}
