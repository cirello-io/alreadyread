// Copyright 2018 github.com/ucirello
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

package actions

import (
	"net/url"

	"cirello.io/alreadyread/pkg/errors"
	"cirello.io/alreadyread/pkg/models"
	"cirello.io/alreadyread/pkg/net"
	"github.com/jmoiron/sqlx"
)

// AddBookmarkByURL reads a URL and inserts its bookmark into the database.
func AddBookmarkByURL(db *sqlx.DB, u string) error {
	if _, err := url.Parse(u); err != nil {
		return errors.Invalidf(err, "invalid URL")
	}

	b := net.CheckLink(&models.Bookmark{
		URL: u,
	})

	_, err := models.NewBookmarkDAO(db).Insert(b)
	return errors.Internal(err)
}

// AddBookmark stores one bookmark into the database.
func AddBookmark(db *sqlx.DB, b *models.Bookmark) error {
	if _, err := url.Parse(b.URL); err != nil {
		return errors.Invalidf(err, "invalid URL")
	}

	b = net.CheckLink(b)
	b, err := models.NewBookmarkDAO(db).Insert(b)
	if err != nil {
		return errors.Internal(err)
	}
	return nil
}
