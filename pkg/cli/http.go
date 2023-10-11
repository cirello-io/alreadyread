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
	"net"
	"net/http"

	"cirello.io/alreadyread/pkg/tasks"
	"cirello.io/alreadyread/pkg/web"
	"github.com/urfave/cli"
)

func (c *commands) httpMode() cli.Command {
	return cli.Command{
		Name:        "http",
		Aliases:     []string{"serve"},
		Usage:       "http mode",
		Description: "starts alreadyread web server",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "bind",
				Value:  ":8080",
				EnvVar: "ALREADYREAD_LISTEN",
			},
		},
		Action: func(ctx *cli.Context) error {
			lHTTP, err := net.Listen("tcp", ctx.String("bind"))
			if err != nil {
				return cliError(fmt.Errorf("cannot bind port: %w", err))
			}
			tasks.Run(c.db)
			srv := web.New(c.db)
			if err := http.Serve(lHTTP, srv); err != nil {
				return cliError(err)
			}
			return nil
		},
	}
}
