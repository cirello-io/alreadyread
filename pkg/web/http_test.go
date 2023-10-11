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

package web

import "testing"

func Test_extractID(t *testing.T) {
	tests := []struct {
		root    string
		url     string
		want    int64
		wantErr bool
	}{
		{"", "/bookmarks/1", 0, true},
		{"/bookmarks", "", 0, true},
		{"/bookmarks", "/bookmarks/1", 1, false},
		{"/bookmarks", "/bookmarks/1/2", 1, false},
		{"/bookmarks", "/bookmarks", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.root+tt.url, func(t *testing.T) {
			got, err := extractID(tt.root, tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractID() = %v, want %v", got, tt.want)
			}
		})
	}
}
