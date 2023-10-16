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
	"testing"

	"cirello.io/alreadyread/pkg/bookmarks"
	"cirello.io/alreadyread/pkg/db"
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

func TestRepository_Bootstrap(t *testing.T) {
	b := New(newConn(t))
	if err := b.Bootstrap(); err != nil {
		t.Error("unexpected error found:", err)
	}
	if err := b.Bootstrap(); err != nil {
		t.Error("unexpected error found (bootstrap should be idempotent):", err)
	}
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
