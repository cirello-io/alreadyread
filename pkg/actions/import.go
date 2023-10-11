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
	"bufio"
	"io"
	"net/http"

	"cirello.io/alreadyread/pkg/errors"
	"cirello.io/alreadyread/pkg/models"
	"github.com/jmoiron/sqlx"
)

// ImportBookmarks takes a reader that has one URL per row and inserts them into
// DB.
func ImportBookmarks(db *sqlx.DB, r io.Reader) error {
	bookmarkDAO := models.NewBookmarkDAO(db)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if _, err := bookmarkDAO.Insert(&models.Bookmark{
			URL:              scanner.Text(),
			LastStatusCode:   http.StatusOK,
			LastStatusCheck:  0,
			LastStatusReason: "",
		}); err != nil {
			return errors.Internal(err)
		}
	}
	if err := scanner.Err(); err != nil {
		return errors.Internalf(err, "reading input")
	}
	return nil
}
