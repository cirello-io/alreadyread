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

package tasks

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"cirello.io/alreadyread/pkg/bookmarks/sqliterepo"
	"cirello.io/alreadyread/pkg/db"
	"github.com/DATA-DOG/go-sqlmock"
)

func Test_vacuum(t *testing.T) {
	t.Run("badDB", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal("cannot create mock:", err)
		}
		errDB := errors.New("bad DB")
		mock.ExpectExec("VACUUM").WillReturnError(errDB)
		if err := vacuum(context.TODO(), db); !errors.Is(err, errDB) {
			t.Fatal("unexpected error", err)
		}
	})
	t.Run("good", func(t *testing.T) {
		db := newDB(t)
		if err := vacuum(context.TODO(), db); err != nil {
			t.Fatal("unexpected error:", err)
		}
	})
}

func Test_restorePostponedLinks(t *testing.T) {
	t.Run("badDB", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal("cannot create mock:", err)
		}
		errDB := errors.New("bad DB")
		mock.ExpectExec("UPDATE bookmarks").WillReturnError(errDB)
		if err := restorePostponedLinks(context.TODO(), db); !errors.Is(err, errDB) {
			t.Fatal("unexpected error", err)
		}
	})
	t.Run("good", func(t *testing.T) {
		db := newDB(t)
		if _, err := db.Exec("INSERT INTO bookmarks (id, inbox, title, created_at) VALUES (1,2,'title',$1)", time.Now()); err != nil {
			t.Fatal("cannot insert row:", err)
		}
		if err := restorePostponedLinks(context.TODO(), db); err != nil {
			t.Fatal("unexpected error:", err)
		}
		row := db.QueryRow("SELECT count(*) FROM bookmarks WHERE inbox = 1")
		var rowCount int
		if err := row.Scan(&rowCount); err != nil {
			t.Fatal("cannot parse row:", err)
		}
		if rowCount != 1 {
			t.Fatal("unexpected row count:", rowCount)
		}
	})
}

func newDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := db.Connect(db.Config{Filename: "file::memory:?cache=shared"})
	if err != nil {
		t.Fatal("cannot create in memory SQLite:", err)
	}
	t.Cleanup(func() { conn.Close() })
	repository := sqliterepo.New(conn)
	if err := repository.Bootstrap(); err != nil {
		t.Fatal("cannot prepare tables:", err)
	}
	return conn
}

func TestRun(t *testing.T) {
	t.Skip("test is incomplete, perhaps the architecture is wrong")
	db := newDB(t)
	if _, err := db.Exec("INSERT INTO bookmarks (id, inbox, title, created_at) VALUES (1,2,'title',$1)", time.Now()); err != nil {
		t.Fatal("cannot insert row:", err)
	}
	got := Run(db)
	if got.Restart == nil || got.Start == nil || got.Shutdown == nil {
		t.Error("incomplete tree definition")
	}
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go func() {
		time.Sleep(5 * time.Second)
		cancel()
	}()
	err := got.Start(ctx)
	t.Log(err)
}
