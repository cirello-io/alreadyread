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

package bookmarks

import "testing"

func TestParseInbox(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name    string
		args    args
		want    Inbox
		wantErr bool
	}{
		{"invalid", args{"invalid"}, Read, true},
		{"read", args{"read"}, Read, false},
		{"new", args{"new"}, NewLink, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInbox(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInbox() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseInbox() = %v, want %v", got, tt.want)
			}
		})
	}
}
