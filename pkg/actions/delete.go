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
	"cirello.io/alreadyread/pkg/errors"
	"cirello.io/alreadyread/pkg/models"
	"github.com/jmoiron/sqlx"
)

// DeleteBookmark deletes one bookmark from the database.
func DeleteBookmark(db *sqlx.DB, b *models.Bookmark) error {
	err := models.NewBookmarkDAO(db).Delete(b)
	if err != nil {
		return errors.Internal(err)
	}
	return nil
}

// DeleteBookmarkByID deletes one bookmar, by ID, from the database.
func DeleteBookmarkByID(db *sqlx.DB, id int64) error {
	bookmark, err := models.NewBookmarkDAO(db).GetByID(id)
	if err != nil {
		return errors.Internal(err)
	}
	return DeleteBookmark(db, bookmark)
}
