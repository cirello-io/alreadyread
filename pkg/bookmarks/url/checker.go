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

package url

import (
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

//go:generate moq -out httpGetter_mocks_test.go . httpGetter
type httpGetter interface {
	Get(url string) (resp *http.Response, err error)
}

type Checker struct {
	timeNow    func() time.Time
	httpClient httpGetter
}

func NewChecker() *Checker {
	return &Checker{
		timeNow:    time.Now,
		httpClient: http.DefaultClient,
	}
}

// Check dials bookmark URL and updates its state with the errors if any.
func (u *Checker) Check(url, originalTitle string) (title string, when int64, code int64, reason string) {
	title = originalTitle
	res, err := u.httpClient.Get(url)
	if err != nil {
		return originalTitle, u.timeNow().Unix(), http.StatusServiceUnavailable, err.Error()
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return originalTitle, u.timeNow().Unix(), int64(res.StatusCode), http.StatusText(res.StatusCode)
	}
	isHTML := strings.Contains(res.Header.Get("Content-Type"), "text/html")
	if originalTitle != "" || !isHTML {
		return originalTitle, u.timeNow().Unix(), int64(res.StatusCode), http.StatusText(res.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err == nil {
		doc.Find("HEAD>TITLE").Each(func(i int, s *goquery.Selection) {
			title = s.Text()
		})
	}
	return title, u.timeNow().Unix(), int64(res.StatusCode), ""
}

func (u *Checker) Title(url string) string {
	title, _, _, _ := u.Check(url, "")
	return title
}
