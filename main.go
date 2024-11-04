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
	"github.com/adhocore/gronx"
	_ "modernc.org/sqlite" // SQLite3 driver
)

var (
	dbFN           = flag.String("db", envOrDefault("ALREADYREAD_DB", "bookmarks.db"), "database filename")
	bind           = flag.String("bind", envOrDefault("ALREADYREAD_LISTEN", ":8080"), "bind address for the server")
	allowedOrigins = flag.String("allowedOrigins", envOrDefault("ALREADYREAD_ALLOWEDORIGINS", "localhost:8080"), "comma-separated value for allowed origins")
	scanDeadLinks  = flag.Bool("scanDeadLinks", false, "scan dead links")
)

func main() {
	log.SetPrefix("")
	log.SetFlags(0)
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	db, err := sql.Open("sqlite", *dbFN)
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
	if *scanDeadLinks {
		err := bookmarks.RefreshExpiredLinks(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("done")
		return
	}

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
					t, _ := gronx.NextTickAfter("0 */6 * * *", time.Now(), false)
					select {
					case <-time.After(time.Until(t)):
						return err
					case <-ctx.Done():
						return ctx.Err()
					}
				},
				Shutdown: oversight.Infinity(),
			},
			oversight.ChildProcessSpecification{
				Name:    "sqliteRestorePostponedLinks",
				Restart: oversight.Permanent(),
				Start: func(ctx context.Context) error {
					err := bookmarks.RestorePostponedLinks(ctx)
					t, _ := gronx.NextTickAfter("15 */6 * * *", time.Now(), false)
					select {
					case <-time.After(time.Until(t)):
						return err
					case <-ctx.Done():
						return ctx.Err()
					}
				},
				Shutdown: oversight.Infinity(),
			},
			oversight.ChildProcessSpecification{
				Name:    "HTTP",
				Restart: oversight.Permanent(),
				Start: func(ctx context.Context) error {
					srv := &http.Server{
						Handler: webserver,
					}
					go func() {
						<-ctx.Done()
						err := srv.Shutdown(context.Background())
						if err != nil {
							log.Println("HTTP server shutdown error:", err)
						}
					}()
					if err := srv.Serve(lHTTP); err != nil {
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
