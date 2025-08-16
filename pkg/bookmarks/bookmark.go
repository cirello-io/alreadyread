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

package bookmarks // import "cirello.io/alreadyread/pkg/bookmarks"

import (
	"fmt"
	"time"
)

// Bookmark stores the basic information of a web URL.
type Bookmark struct {
	ID               int64     `db:"id" json:"id"`
	URL              string    `db:"url" json:"url"`
	LastStatusCode   int64     `db:"last_status_code" json:"last_status_code"`
	LastStatusCheck  int64     `db:"last_status_check" json:"last_status_check"`
	LastStatusReason string    `db:"last_status_reason" json:"last_status_reason"`
	Title            string    `db:"title" json:"title"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	Inbox            Inbox     `db:"inbox" json:"inbox"`
	Description      string    `db:"description" json:"description"`
	BumpDate         time.Time `db:"bump_date" json:"bump_date"`

	Host string `db:"-" json:"host"`
}

type Inbox int64

const (
	Read Inbox = iota
	NewLink
)

func ParseInbox(v string) (Inbox, error) {
	switch v {
	case "read":
		return Read, nil
	case "new":
		return NewLink, nil
	default:
		return 0, fmt.Errorf("invalid inbox status: %s", v)
	}
}
