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

package main // import "cirello.io/alreadyread"

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"cirello.io/alreadyread/pkg/bookmarks"
	"cirello.io/alreadyread/pkg/bookmarks/sqliterepo"
	"cirello.io/alreadyread/pkg/bookmarks/url"
	"cirello.io/alreadyread/pkg/db"
	"cirello.io/alreadyread/pkg/tasks"
	"cirello.io/alreadyread/pkg/web"
	"cirello.io/oversight"
)

var (
	dbFN           = flag.String("db", envOrDefault("ALREADYREAD_DB", "bookmarks.db"), "database filename")
	bind           = flag.String("bind", envOrDefault("ALREADYREAD_LISTEN", ":8080"), "bind address for the server")
	allowedOrigins = flag.String("allowedOrigins", envOrDefault("ALREADYREAD_ALLOWEDORIGINS", "localhost:8080"), "comma-separated value for allowed origins")
)

func main() {
	log.SetPrefix("")
	log.SetFlags(0)
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	db, err := db.Connect(db.Config{Filename: *dbFN})
	if err != nil {
		log.Fatal(err)
	}
	lHTTP, err := net.Listen("tcp", *bind)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot bind port: %w", err))
	}
	go func() {
		<-ctx.Done()
		lHTTP.Close()
	}()
	svr := oversight.New(
		oversight.WithLogger(log.Default()),
		oversight.WithRestartStrategy(oversight.OneForAll()),
		oversight.NeverHalt(),
	)
	svr.Add(tasks.Run(db))
	svr.Add(webserver(lHTTP, db))
	svr.Start(ctx)
}

func webserver(listener net.Listener, db *sql.DB) oversight.ChildProcessSpecification {
	bookmarks := bookmarks.New(sqliterepo.New(db), url.NewChecker())
	srv := web.New(bookmarks, url.NewChecker(), strings.Split(*allowedOrigins, ","))
	return oversight.ChildProcessSpecification{
		Name:    "HTTP",
		Restart: oversight.Permanent(),
		Start: func(ctx context.Context) error {
			if err := http.Serve(listener, srv); err != nil {
				return fmt.Errorf("HTTP server error: %w", err)
			}
			return nil
		},
		Shutdown: oversight.Infinity(),
	}
}

func envOrDefault(name string, defaultValue string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return defaultValue
}
