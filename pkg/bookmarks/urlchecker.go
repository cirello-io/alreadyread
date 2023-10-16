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

//go:generate moq -out urlchecker_mocks_test.go . URLChecker
//go:generate moq -pkg web -out ../web/urlchecker_mocks_test.go . URLChecker
type URLChecker interface {
	Check(url, originalTitle string) (title string, when int64, code int64, reason string)
	Title(url string) (title string)
}
