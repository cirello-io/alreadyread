// Copyright 2019 github.com/ucirello
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

	"cirello.io/alreadyread/pkg/models"
	"github.com/jmoiron/sqlx"
)

// UpdateInbox sets the inbox status of a bookmark.
func UpdateInbox(db *sqlx.DB, id int64, inbox string) error {
	parsedInbox, err := models.ParseInbox(inbox)
	if err != nil {
		return fmt.Errorf("cannot parse inbox: %w", err)
	}
	dao := models.NewBookmarkDAO(db)
	b, err := dao.GetByID(id)
	if err != nil {
		return fmt.Errorf("cannot find bookmark: %w", err)
	}
	b.Inbox = parsedInbox
	if err := dao.Update(b); err != nil {
		return fmt.Errorf("cannot store bookmark: %w", err)
	}
	return nil
}
