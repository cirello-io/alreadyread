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

package actions

import (
	"fmt"
	"net/url"

	"cirello.io/alreadyread/pkg/models"
	"cirello.io/alreadyread/pkg/net"
	"github.com/jmoiron/sqlx"
)

// AddBookmark stores one bookmark into the database.
func AddBookmark(db *sqlx.DB, b *models.Bookmark) (*models.Bookmark, error) {
	if _, err := url.Parse(b.URL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	b = net.CheckLink(b)
	if _, err := models.NewBookmarkDAO(db).Insert(b); err != nil {
		return nil, fmt.Errorf("cannot insert bookmark: %w", err)
	}
	return b, nil
}
