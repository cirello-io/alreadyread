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
	"net"
	"net/http"

	"cirello.io/alreadyread/pkg/errors"
	"cirello.io/alreadyread/pkg/pubsub"
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
				EnvVar: "BOOKMARKD_LISTEN",
			},
		},
		Action: func(ctx *cli.Context) error {
			lHTTP, err := net.Listen("tcp", ctx.String("bind"))
			if err != nil {
				return errors.Errorf(err, "cannot bind port")
			}
			broker := pubsub.New()
			tasks.Run(c.db)
			srv, err := web.New(c.db, broker)
			if err != nil {
				return errors.E(err)
			}
			err = http.Serve(lHTTP, srv)
			return errors.E(err)
		},
	}
}
