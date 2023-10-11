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
	"cirello.io/alreadyread/pkg/errors"
	"cirello.io/alreadyread/pkg/models"
	"github.com/jmoiron/sqlx"
)

// MarkBookmarkAsRead moves a bookmark out of inbox.
func MarkBookmarkAsRead(db *sqlx.DB, id int64) error {
	dao := models.NewBookmarkDAO(db)
	b, err := dao.GetByID(id)
	if err != nil {
		return errors.Internalf(err, "cannot find bookmark")
	}
	b.Inbox = 0
	if err := dao.Update(b); err != nil {
		return errors.Internalf(err, "cannot update bookmarkd")
	}
	return nil
}
