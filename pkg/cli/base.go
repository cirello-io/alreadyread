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

package cli // import "cirello.io/alreadyread/pkg/cli"

import (
	"log"
	"os"
	"sort"
	"strings"

	"cirello.io/alreadyread/pkg/bookmarks/sqliterepo"
	"github.com/jmoiron/sqlx"
	"github.com/urfave/cli"
)

type commands struct {
	db *sqlx.DB
}

func (c *commands) bootstrap(*cli.Context) error {
	return sqliterepo.New(c.db).Bootstrap()
}

// Run executes the application in CLI mode.
func Run(db *sqlx.DB) {
	app := cli.NewApp()
	app.Name = "alreadyread"
	app.Usage = "bookmarks manager"
	app.Version = "0.0.1"

	cmds := &commands{
		db: db,
	}
	app.Before = cmds.bootstrap
	app.Commands = []cli.Command{
		cmds.httpMode(),
		cmds.listBookmarks(),
	}
	sort.Slice(app.Commands, func(i, j int) bool {
		return strings.Compare(app.Commands[i].Name, app.Commands[j].Name) < 0
	})
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func cliError(err error) error {
	if err == nil {
		return nil
	}
	return cli.NewExitError(err.Error(), 1)
}
