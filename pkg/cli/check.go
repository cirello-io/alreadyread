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

package cli

import (
	"cirello.io/alreadyread/pkg/actions"
	"cirello.io/alreadyread/pkg/errors"
	"github.com/urfave/cli"
)

func (c *commands) checkBookmarks() cli.Command {
	return cli.Command{
		Name:        "check",
		Usage:       "check bookmarks",
		Description: "look for broken links",
		Action: func(ctx *cli.Context) error {
			err := actions.CheckBookmarks(c.db)
			if err != nil {
				return cliError(errors.Errorf(err, "cannot check bookmarks"))
			}

			return nil
		},
	}
}
