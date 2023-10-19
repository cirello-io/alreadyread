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
	"time"

	"cirello.io/alreadyread/pkg/bookmarks"
	"cirello.io/alreadyread/pkg/bookmarks/sqliterepo"
	"cirello.io/alreadyread/pkg/bookmarks/url"
	"cirello.io/alreadyread/pkg/web"
	"cirello.io/oversight"
	_ "github.com/mattn/go-sqlite3" // SQLite3 driver
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

	db, err := sql.Open("sqlite3", *dbFN)
	if err != nil {
		log.Println(err)
		return
	}
	lHTTP, err := net.Listen("tcp", *bind)
	if err != nil {
		log.Println("cannot bind port:", err)
		return
	}
	go func() {
		<-ctx.Done()
		lHTTP.Close()
	}()

	repository := sqliterepo.New(db)
	if err := repository.Bootstrap(); err != nil {
		log.Println("cannot bootstrap database:", err)
		return
	}
	bookmarks := bookmarks.New(repository, url.NewChecker())
	webserver := web.New(bookmarks, url.NewChecker(), strings.Split(*allowedOrigins, ","))

	svr := oversight.New(
		oversight.WithLogger(log.Default()),
		oversight.WithRestartStrategy(oversight.OneForOne()),
		oversight.NeverHalt(),
		oversight.Process(
			oversight.ChildProcessSpecification{
				Name:    "sqliteVacuum",
				Restart: oversight.Permanent(),
				Start: func(ctx context.Context) error {
					err := repository.Vacuum(ctx)
					time.Sleep(6 * time.Hour)
					return err
				},
				Shutdown: oversight.Infinity(),
			},
			oversight.ChildProcessSpecification{
				Name:    "sqliteRestorePostponedLinks",
				Restart: oversight.Permanent(),
				Start: func(ctx context.Context) error {
					err := repository.RestorePostponedLinks(ctx)
					time.Sleep(6 * time.Hour)
					return err
				},
				Shutdown: oversight.Infinity(),
			},
			oversight.ChildProcessSpecification{
				Name:    "refreshExpiredLinks",
				Restart: oversight.Permanent(),
				Start: func(ctx context.Context) error {
					err := bookmarks.RefreshExpiredLinks(ctx)
					time.Sleep(6 * time.Hour)
					return err
				},
				Shutdown: oversight.Infinity(),
			},
			oversight.ChildProcessSpecification{
				Name:    "HTTP",
				Restart: oversight.Permanent(),
				Start: func(ctx context.Context) error {
					if err := http.Serve(lHTTP, webserver); err != nil {
						return fmt.Errorf("HTTP server error: %w", err)
					}
					return nil
				},
				Shutdown: oversight.Infinity(),
			},
		),
	)

	if err := svr.Start(ctx); err != nil {
		log.Println("oversight tree error:", err)
		return
	}
}

func envOrDefault(name string, defaultValue string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return defaultValue
}
