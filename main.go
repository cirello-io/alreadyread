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
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"cirello.io/alreadyread/pkg/bookmarks"
	"cirello.io/alreadyread/pkg/bookmarks/sqliterepo"
	"cirello.io/alreadyread/pkg/bookmarks/url"
	"cirello.io/alreadyread/pkg/db"
	"cirello.io/alreadyread/pkg/tasks"
	"cirello.io/alreadyread/pkg/web"
)

var (
	bind           = flag.String("bind", envOrDefault("ALREADYREAD_LISTEN", ":8080"), "bind address for the server")
	allowedOrigins = flag.String("allowedOrigins", envOrDefault("ALREADYREAD_ALLOWEDORIGINS", "localhost:8080"), "comma-separated value for allowed origins")
)

func main() {
	log.SetPrefix("")
	log.SetFlags(0)
	flag.Parse()

	fn := "bookmarks.db"
	if envFn := os.Getenv("ALREADYREAD_DB"); envFn != "" {
		fn = envFn
	}
	db, err := db.Connect(db.Config{Filename: fn})
	if err != nil {
		log.Fatal(err)
	}

	lHTTP, err := net.Listen("tcp", *bind)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot bind port: %w", err))
	}
	tasks.Run(db)
	bookmarks := bookmarks.New(sqliterepo.New(db), url.NewChecker())
	srv := web.New(bookmarks, url.NewChecker(), strings.Split(*allowedOrigins, ","))
	if err := http.Serve(lHTTP, srv); err != nil {
		log.Fatal(err)
	}
}

func envOrDefault(name string, defaultValue string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return defaultValue
}
