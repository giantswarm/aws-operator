// Copyright 2018 Richard Lehane. All rights reserved.
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
package crock32

// Sortable makes a slice of crock32 strings sortable numerically
type Sortable []string

func (s Sortable) Len() int { return len(s) }
func (s Sortable) Less(i, j int) bool {
	ii, _ := Decode(s[i])
	jj, _ := Decode(s[j])
	return ii < jj
}
func (s Sortable) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
