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
	"testing"
	"time"
)

func TestCheckLink(t *testing.T) {
	now := func() time.Time {
		return time.Unix(0, 0)
	}
	checker := NewChecker()
	checker.timeNow = now

	tests := []struct {
		name       string
		url        string
		title      string
		wantURL    string
		wantTitle  string
		wantCode   int64
		wantWhen   int64
		wantReason string
	}{
		{
			name:       "404",
			url:        "http://example.com/404",
			wantURL:    "http://example.com/404",
			wantTitle:  "",
			wantCode:   404,
			wantWhen:   now().Unix(),
			wantReason: "Not Found",
		},
		{
			name:       "200",
			url:        "http://example.com/",
			wantURL:    "http://example.com/",
			wantTitle:  "Example Domain",
			wantCode:   200,
			wantWhen:   now().Unix(),
			wantReason: "",
		},
		{
			name:       "Custom Title",
			url:        "http://example.com/",
			title:      "Custom Title",
			wantURL:    "http://example.com/",
			wantTitle:  "Custom Title",
			wantCode:   200,
			wantWhen:   now().Unix(),
			wantReason: "OK",
		},
		{
			name:       "invalid URL",
			url:        "invalid-url",
			wantURL:    "invalid-url",
			wantTitle:  "",
			wantCode:   0,
			wantWhen:   now().Unix(),
			wantReason: "Get \"invalid-url\": unsupported protocol scheme \"\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTitle, gotWhen, gotCode, gotReason := checker.Check(tt.url, tt.title)
			if gotTitle != tt.wantTitle {
				t.Errorf("%s CheckLink().Title = %v, want %v", tt.name, gotTitle, tt.wantTitle)
			}
			if gotWhen != tt.wantWhen {
				t.Errorf("%s CheckLink().When = %v, want %v", tt.name, gotWhen, tt.wantWhen)
			}
			if gotCode != tt.wantCode {
				t.Errorf("%s CheckLink().Code = %v, want %v", tt.name, gotCode, tt.wantCode)
			}
			if gotReason != tt.wantReason {
				t.Errorf("%s CheckLink().Reason = %v, want %v", tt.name, gotReason, tt.wantReason)
			}

		})
	}
}

func TestContentExtraction(t *testing.T) {
	checker := NewChecker()
	checker.timeNow = func() time.Time {
		return time.Unix(0, 0)
	}
	title, _, _, _ := checker.Check("https://www.example.org", "")
	if title != "Example Domain" {
		t.Fatal("cannot extract HTML title")
	}
}
