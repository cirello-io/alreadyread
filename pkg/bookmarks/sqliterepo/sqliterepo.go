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
	"fmt"
	"net/url"
	"time"

	"cirello.io/alreadyread/pkg/bookmarks"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

// New instanties a SQLite based repository.
func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (b *Repository) Bootstrap() error {
	cmds := []string{
		`create table if not exists bookmarks (
			id integer primary key autoincrement,
			url text,
			last_status_code int,
			last_status_check int,
			last_status_reason text,
			title bigtext not null,
			created_at datetime not null,
			inbox int not null default 0
		);
		`,
		`create index if not exists bookmarks_last_status_code  on bookmarks (last_status_code)`,
		`create index if not exists bookmarks_last_status_check on bookmarks (last_status_check)`,
		`create index if not exists bookmarks_created_at on bookmarks (created_at)`,
		`create index if not exists bookmarks_inbox on bookmarks (inbox)`,
	}

	for _, cmd := range cmds {
		_, err := b.db.Exec(cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Repository) All() ([]*bookmarks.Bookmark, error) {
	var bookmarks []*bookmarks.Bookmark
	err := b.db.Select(&bookmarks, `
		SELECT
			*

		FROM
			bookmarks

		ORDER BY
			CASE
				WHEN last_status_code = 0 THEN 999
				ELSE last_status_code
			END ASC,
			created_at DESC,
			id DESC
	`)

	for _, b := range bookmarks {
		u, err := url.Parse(b.URL)
		if err == nil {
			b.Host = u.Host
		}
	}
	if err != nil {
		return nil, err
	}
	return bookmarks, nil
}

func (b *Repository) Expired() ([]*bookmarks.Bookmark, error) {
	var bookmarks []*bookmarks.Bookmark
	const week = 7 * 24 * time.Hour
	deadline := time.Now().Add(-week).Unix()
	err := b.db.Select(&bookmarks, `
		SELECT
			*
		FROM
			bookmarks
		WHERE
			last_status_code = 200
			AND last_status_check <= $1
	`, deadline)
	for _, b := range bookmarks {
		u, err := url.Parse(b.URL)
		if err == nil {
			b.Host = u.Host
		}
	}
	if err != nil {
		return nil, err
	}
	return bookmarks, nil
}

func (b *Repository) Invalid() ([]*bookmarks.Bookmark, error) {
	var bookmarks []*bookmarks.Bookmark
	err := b.db.Select(&bookmarks, `
		SELECT
			*
		FROM
			bookmarks
		WHERE
			last_status_code != 200
	`)
	if err != nil {
		return nil, err
	}
	for _, b := range bookmarks {
		u, err := url.Parse(b.URL)
		if err == nil {
			b.Host = u.Host
		}
	}
	return bookmarks, nil
}

func (b *Repository) Insert(bookmark *bookmarks.Bookmark) error {
	bookmark.CreatedAt = time.Now()
	bookmark.Inbox = 1
	result, err := b.db.NamedExec(`
		INSERT INTO bookmarks
		(url, last_status_code, last_status_check, last_status_reason, title, created_at, inbox)
		VALUES
		(:url, :last_status_code, :last_status_check, :last_status_reason, :title, :created_at, :inbox)
	`, bookmark)
	if err != nil {
		return fmt.Errorf("cannot insert row: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("cannot load last inserted ID: %w", err)
	}
	err = b.db.Get(bookmark, `
		SELECT
			*
		FROM
			bookmarks
		WHERE
			id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("cannot reload inserted row: %w", err)
	}
	u, err := url.Parse(bookmark.URL)
	if err == nil {
		bookmark.Host = u.Host
	}
	return nil
}

func (b *Repository) GetByID(id int64) (*bookmarks.Bookmark, error) {
	bookmark := &bookmarks.Bookmark{}
	err := b.db.Get(bookmark, `
	SELECT
		*
	FROM
		bookmarks
	WHERE
		id = $1
	`, id)
	if err != nil {
		return nil, fmt.Errorf("cannot find row: %w", err)
	}
	u, err := url.Parse(bookmark.URL)
	if err != nil {
		return bookmark, fmt.Errorf("cannot parse URL: %w", err)
	}
	bookmark.Host = u.Host
	return bookmark, nil
}

func (b *Repository) Update(bookmark *bookmarks.Bookmark) error {
	_, err := b.db.NamedExec(`
		UPDATE bookmarks
		SET
			url = :url,
			last_status_code = :last_status_code,
			last_status_check = :last_status_check,
			last_status_reason = :last_status_reason,
			title = :title,
			inbox = :inbox
		WHERE
			id = :id
	`, bookmark)
	return err
}

func (b *Repository) DeleteByID(id int64) error {
	_, err := b.db.Exec(`DELETE FROM bookmarks WHERE id = $1`, id)
	return err
}
