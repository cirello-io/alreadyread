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

package sqliterepo

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"cirello.io/alreadyread/pkg/bookmarks"
	"cirello.io/alreadyread/pkg/db"
	"github.com/DATA-DOG/go-sqlmock"
)

func newConn(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := db.Connect(db.Config{Filename: ":memory:"})
	if err != nil {
		t.Fatal("cannot create in memory SQLite:", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func setup(t *testing.T) *Repository {
	t.Helper()
	b := New(newConn(t))
	if err := b.Bootstrap(); err != nil {
		t.Fatal("cannot run bootstrap:", err)
	}
	return b
}

func TestRepository_Bootstrap(t *testing.T) {
	t.Run("badDB", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal("cannot create mock:", err)
		}
		errDB := errors.New("bad DB")
		mock.ExpectExec("create table").WillReturnError(errDB)
		b := New(db)
		if err := b.Bootstrap(); !errors.Is(err, errDB) {
			t.Error("expected error missing: ", err)
		}
	})
	t.Run("good", func(t *testing.T) {
		b := New(newConn(t))
		if err := b.Bootstrap(); err != nil {
			t.Error("unexpected error found:", err)
		}
		if err := b.Bootstrap(); err != nil {
			t.Error("unexpected error found (bootstrap should be idempotent):", err)
		}
	})
}

func TestRepository_basicCycle(t *testing.T) {
	b := New(newConn(t))
	if err := b.Bootstrap(); err != nil {
		t.Fatal("unexpected error found:", err)
	}
	inserted := &bookmarks.Bookmark{
		URL:   "https://example.com",
		Title: "title",
		Inbox: bookmarks.NewLink,
	}
	if err := b.Insert(inserted); err != nil {
		t.Fatal("cannot insert bookmark:", err)
	}
	t.Log("bookmark.ID:", inserted.ID)
	loaded, err := b.GetByID(inserted.ID)
	if err != nil {
		t.Fatal("cannot load bookmark:", err)
	}
	isEqual := inserted.ID == loaded.ID &&
		inserted.URL == loaded.URL &&
		inserted.Title == loaded.Title
	if !isEqual {
		t.Fatalf("inserted and loaded rows do not match\n%#v\n%#v", inserted, loaded)
	}
	updated := &bookmarks.Bookmark{
		ID:    loaded.ID,
		Title: "new-title",
		URL:   "https://newurl.com",
		Inbox: bookmarks.NewLink,
	}
	if err := b.Update(updated); err != nil {
		t.Fatal("cannot update bookmark:", err)
	}
	inbox, err := b.Inbox()
	if err != nil {
		t.Fatal("cannot load inbox bookmarks:", err)
	}
	if l := len(inbox); l != 1 {
		t.Fatal("unexpected row count:", l)
	}
	isUpdated := inbox[0].ID == updated.ID &&
		inbox[0].Title == updated.Title &&
		inbox[0].URL == updated.URL
	if !isUpdated {
		t.Fatal("failed to update the bookmark")
	}
	if err := b.DeleteByID(inbox[0].ID); err != nil {
		t.Fatal("cannot delete bookmark:", err)
	}
	all, err := b.All()
	if err != nil {
		t.Fatal("cannot load all bookmarks:", err)
	}
	if l := len(all); l != 0 {
		t.Fatal("unexpected number of bookmarks:", l)
	}
}

func TestRepository_scanRows(t *testing.T) {
	t.Run("badRows", func(t *testing.T) {
		errRows := errors.New("bad rows")
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal("cannot create mock:", err)
		}
		mock.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows([]string{"id"}).CloseError(errRows),
		)
		b := New(db)
		rows, err := b.db.Query(`SELECT id, url, last_status_code, last_status_check, last_status_reason, title, created_at, inbox FROM bookmarks`)
		if err != nil {
			t.Fatal("unexpected query error:", err)
		}
		if _, err := b.scanRows(rows); !errors.Is(err, errRows) {
			t.Fatal("unexpected scan rows error:", err)
		}
	})
	t.Run("badRowScan", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal("cannot create mock:", err)
		}
		mock.ExpectQuery("SELECT").WillReturnRows(
			sqlmock.NewRows([]string{"id", "url", "last_status_code", "last_status_check", "last_status_reason", "title", "created_at", "inbox"}).
				// good row
				AddRow(1, "http://example.com", 200, time.Now().Unix(), "reason", "title", time.Now(), 0).
				// bad row, will trip row scanner.
				AddRow(2, "http://example.com", 200, time.Now(), "reason", "title", time.Now(), 0),
		)
		b := New(db)
		rows, err := b.db.Query(`SELECT id, url, last_status_code, last_status_check, last_status_reason, title, created_at, inbox FROM bookmarks`)
		if err != nil {
			t.Fatal("unexpected query error:", err)
		}
		if _, err := b.scanRows(rows); err == nil {
			t.Fatal("expected scan row error missing")
		}
	})
}

func TestRepository_scanRow(t *testing.T) {
	errRow := errors.New("bad row")
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal("cannot create mock:", err)
	}
	mock.ExpectQuery("SELECT").WillReturnRows(
		sqlmock.NewRows([]string{"id", "url", "last_status_code", "last_status_check", "last_status_reason", "title", "created_at", "inbox"}).
			AddRow(1, "http://example.com", 200, time.Now(), "reason", "title", time.Now(), 0).
			RowError(0, errRow),
	)
	b := New(db)
	row := b.db.QueryRow(`
	SELECT id, url, last_status_code, last_status_check, last_status_reason, title, created_at, inbox FROM bookmarks
	`)
	if _, err := b.scanRow(row); !errors.Is(err, errRow) {
		t.Fatal("unexpected error:", err)
	}
}

func TestRepository_Insert(t *testing.T) {
	t.Run("badDB", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal("cannot create mock:", err)
		}
		errDB := errors.New("bad DB")
		mock.ExpectExec("INSERT INTO bookmarks").WillReturnError(errDB)
		if err := New(db).Insert(&bookmarks.Bookmark{}); !errors.Is(err, errDB) {
			t.Error("expected error missing: ", err)
		}
	})
	t.Run("badLastInsertID", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal("cannot create mock:", err)
		}
		errResult := errors.New("bad result")
		mock.ExpectExec("INSERT INTO bookmarks").WillReturnResult(sqlmock.NewErrorResult(errResult))
		if err := New(db).Insert(&bookmarks.Bookmark{}); !errors.Is(err, errResult) {
			t.Error("expected error missing: ", err)
		}
	})
	t.Run("good", func(t *testing.T) {
		bookmark := &bookmarks.Bookmark{URL: "http://example.com"}
		if err := setup(t).Insert(bookmark); err != nil {
			t.Error("could not insert bookmark:", err)
		}
		if bookmark.ID == 0 {
			t.Error("did not update bookmark ID")
		}
	})
}