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
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"cirello.io/alreadyread/pkg/bookmarks"
)

type Repository struct {
	db *sql.DB
}

// New instanties a SQLite based repository.
func New(db *sql.DB) *Repository {
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
		`alter table bookmarks add column description bigtext not null default ''`,
	}
	var version int
	row := b.db.QueryRow("PRAGMA user_version;")
	if err := row.Scan(&version); err != nil {
		return err
	}
	if version == 0 {
		version = -1
	}
	for stmt, cmd := range cmds {
		if version >= stmt {
			continue
		}
		_, err := b.db.Exec(cmd)
		if err != nil {
			return fmt.Errorf("apply migration: %w", err)
		}
		_, err = b.db.Exec(fmt.Sprintf("PRAGMA user_version = %d", stmt))
		if err != nil {
			return fmt.Errorf("update migration index: %w", err)
		}
	}
	return nil
}

func (b *Repository) scanRows(rows *sql.Rows) ([]*bookmarks.Bookmark, error) {
	var list []*bookmarks.Bookmark
	for rows.Next() {
		bookmark, err := b.scanRow(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, bookmark)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (b *Repository) scanRow(row interface{ Scan(dest ...any) error }) (*bookmarks.Bookmark, error) {
	bookmark := &bookmarks.Bookmark{}
	if err := row.Scan(&bookmark.ID, &bookmark.URL, &bookmark.LastStatusCode, &bookmark.LastStatusCheck, &bookmark.LastStatusReason, &bookmark.Title, &bookmark.CreatedAt, &bookmark.Inbox, &bookmark.Description); err != nil {
		return nil, err
	}
	u, err := url.Parse(bookmark.URL)
	if err == nil {
		bookmark.Host = u.Host
	}
	return bookmark, nil
}

func (b *Repository) Inbox() ([]*bookmarks.Bookmark, error) {
	rows, err := b.db.Query(`SELECT id, url, last_status_code, last_status_check, last_status_reason, title, created_at, inbox, description FROM bookmarks WHERE inbox = 1 ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	return b.scanRows(rows)
}

func (b *Repository) Duplicated() ([]*bookmarks.Bookmark, error) {
	rows, err := b.db.Query(`SELECT id, url, last_status_code, last_status_check, last_status_reason, title, created_at, inbox, description FROM bookmarks WHERE url IN (SELECT url FROM bookmarks GROUP BY url HAVING count(url) > 1) ORDER BY url, created_at DESC`)
	if err != nil {
		return nil, err
	}
	return b.scanRows(rows)
}

func (b *Repository) Dead() ([]*bookmarks.Bookmark, error) {
	rows, err := b.db.Query(`SELECT id, url, last_status_code, last_status_check, last_status_reason, title, created_at, inbox, description FROM bookmarks WHERE NOT (last_status_code == 200 OR last_status_code == 0) ORDER BY created_at DESC, last_status_code DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	return b.scanRows(rows)
}

func (b *Repository) All() ([]*bookmarks.Bookmark, error) {
	rows, err := b.db.Query(`SELECT id, url, last_status_code, last_status_check, last_status_reason, title, created_at, inbox, description FROM bookmarks ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	return b.scanRows(rows)
}

func (b *Repository) Expired() ([]*bookmarks.Bookmark, error) {
	const week = 7 * 24 * time.Hour
	deadline := time.Now().Add(-week).Unix()
	rows, err := b.db.Query(`SELECT id, url, last_status_code, last_status_check, last_status_reason, title, created_at, inbox, description FROM bookmarks WHERE last_status_code IN (200,0) AND last_status_check <= $1`, deadline)
	if err != nil {
		return nil, err
	}
	return b.scanRows(rows)
}

func (b *Repository) Invalid() ([]*bookmarks.Bookmark, error) {
	rows, err := b.db.Query(`SELECT id, url, last_status_code, last_status_check, last_status_reason, title, created_at, inbox, description FROM bookmarks WHERE last_status_code != 200`)
	if err != nil {
		return nil, err
	}
	return b.scanRows(rows)
}

func (b *Repository) Insert(bookmark *bookmarks.Bookmark) error {
	bookmark.CreatedAt = time.Now()
	bookmark.Inbox = 1
	result, err := b.db.Exec(`
		INSERT INTO bookmarks
		(url, last_status_code, last_status_check, last_status_reason, title, created_at, inbox, description)
		VALUES
		($1, $2, $3, $4, $5, $6, $7, $8)
	`, bookmark.URL, bookmark.LastStatusCode, bookmark.LastStatusCheck, bookmark.LastStatusReason, bookmark.Title, bookmark.CreatedAt, bookmark.Inbox, bookmark.Description)
	if err != nil {
		return fmt.Errorf("cannot insert row: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("cannot load inserted ID: %w", err)
	}
	bookmark.ID = id
	return nil
}

func (b *Repository) GetByID(id int64) (*bookmarks.Bookmark, error) {
	row := b.db.QueryRow(`
	SELECT
		id, url, last_status_code, last_status_check, last_status_reason, title, created_at, inbox, description
	FROM
		bookmarks
	WHERE
		id = $1
	`, id)
	return b.scanRow(row)
}

func (b *Repository) Update(bookmark *bookmarks.Bookmark) error {
	_, err := b.db.Exec(`
		UPDATE bookmarks
		SET
			url = $1,
			last_status_code = $2,
			last_status_check = $3,
			last_status_reason = $4,
			title = $5,
			inbox = $6,
			description = $7
		WHERE
			id = $8
	`, bookmark.URL, bookmark.LastStatusCode, bookmark.LastStatusCheck, bookmark.LastStatusReason, bookmark.Title, bookmark.Inbox, bookmark.Description, bookmark.ID)
	return err
}

func (b *Repository) DeleteByID(id int64) error {
	_, err := b.db.Exec(`DELETE FROM bookmarks WHERE id = $1`, id)
	return err
}

func (b *Repository) Search(term string) ([]*bookmarks.Bookmark, error) {
	explodedTerm := "%" + strings.Join(strings.Split(term, ""), "%") + "%"
	rows, err := b.db.Query(`
		SELECT
			id, url, last_status_code, last_status_check, last_status_reason, title, created_at, inbox, description
		FROM
			bookmarks
		WHERE
			title LIKE $1 COLLATE NOCASE
			OR
			url LIKE $1 COLLATE NOCASE
			OR
			description LIKE $1 COLLATE NOCASE
		ORDER BY
			CASE
				WHEN title = $2 THEN 3
				WHEN title LIKE $3 THEN 2
				WHEN title LIKE $4 THEN 1
				WHEN title LIKE $5 THEN 1
				ELSE 0
			END DESC,
			created_at DESC,
			id DESC
	`, explodedTerm, term, term+"%", "%"+term+"%", "%"+term)
	if err != nil {
		return nil, err
	}
	return b.scanRows(rows)
}

func (b *Repository) Vacuum(ctx context.Context) error {
	if _, err := b.db.ExecContext(ctx, "VACUUM"); err != nil {
		return fmt.Errorf("cannot run vacuum: %w", err)
	}
	return nil
}

func (b *Repository) RestorePostponedLinks(ctx context.Context) error {
	if _, err := b.db.ExecContext(ctx, "UPDATE bookmarks SET inbox = 1 WHERE inbox = 2"); err != nil {
		return fmt.Errorf("cannot run restore rescheduled links: %w", err)
	}
	return nil
}
