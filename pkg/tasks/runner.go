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
	"cirello.io/alreadyread/pkg/bookmarks/url"
	"cirello.io/oversight"
)

type task struct {
	Name      string
	Run       func(context.Context, *sql.DB) error
	Frequency time.Duration
}

var tasks = []task{
	{"check link health", linkHealth, 6 * time.Hour},
	{"vacuum", vacuum, 12 * time.Hour},
	{"restore postponed links", restorePostponedLinks, 6 * time.Hour},
}

// Run executes background maintenance tasks.
func Run(db *sql.DB) oversight.ChildProcessSpecification {
	svr := oversight.New(
		oversight.WithRestartStrategy(oversight.OneForOne()),
		oversight.NeverHalt(),
	)
	for _, t := range tasks {
		t := t
		log.Println("scheduled", t.Name)
		svr.Add(oversight.ChildProcessSpecification{
			Name:    t.Name,
			Restart: oversight.Permanent(),
			Start: func(ctx context.Context) error {
				log.Println(t.Name + ": start")
				if err := t.Run(ctx, db); err != nil {
					log.Println(t.Name+": error ", err)
				}
				log.Println(t.Name + ": done")
				select {
				case <-time.After(t.Frequency):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			},
			Shutdown: oversight.Infinity(),
		})
	}
	return oversight.ChildProcessSpecification{
		Name:     "tasks",
		Restart:  oversight.Permanent(),
		Start:    svr.Start,
		Shutdown: oversight.Infinity(),
	}
}

// linkHealth checks if the expired links are still valid.
func linkHealth(ctx context.Context, db *sql.DB) (err error) {
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
			urlChecker := url.NewChecker()
			for bookmark := range bookmarkCh {
				log.Println("linkHealth:", bookmark.ID, bookmark.URL)
				bookmark.Title, bookmark.LastStatusCheck, bookmark.LastStatusCode, bookmark.LastStatusReason = urlChecker.Check(bookmark.URL, bookmark.Title)
				if err := repository.Update(bookmark); err != nil {
					log.Println(err, "cannot update link during link health check - status OK")
				}
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

	return nil
}

// vacuum executes a SQLite3 vacuum clean up.
func vacuum(ctx context.Context, db *sql.DB) (err error) {
	defer recoverPanic(&err)
	_, err = db.ExecContext(ctx, "VACUUM")
	if err != nil {
		return fmt.Errorf("cannot run vacuum: %w", err)
	}
	return nil
}

// restorePostponedLinks revamp rescheduled links in the inbox.
func restorePostponedLinks(ctx context.Context, db *sql.DB) (err error) {
	defer recoverPanic(&err)
	_, err = db.ExecContext(ctx, "UPDATE bookmarks SET inbox = 1 WHERE inbox = 2")
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
