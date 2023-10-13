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

package cli

import (
	"fmt"
	"text/tabwriter"

	"cirello.io/alreadyread/pkg/bookmarks"
	"cirello.io/alreadyread/pkg/bookmarks/sqliterepo"
	"github.com/urfave/cli"
)

func (c *commands) listBookmarks() cli.Command {
	return cli.Command{
		Name:        "list",
		Usage:       "list bookmarks",
		Description: "list all bookmarks",
		Action: func(ctx *cli.Context) error {
			repository := sqliterepo.New(c.db)
			bookmarks, err := bookmarks.New(repository).All()
			if err != nil {
				return cliError(fmt.Errorf("cannot load bookmarks: %w", err))
			}
			w := tabwriter.NewWriter(ctx.App.Writer, 0, 0, 1, ' ', 0)
			for _, b := range bookmarks {
				fmt.Fprintln(w, b.Title, "\t", b.URL)
			}
			w.Flush()
			return nil
		},
	}
}
