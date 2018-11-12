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

import (
	"math/rand"
	"time"
)

func pseudo(r, t int64) uint64 {
	r = r & 0xFFFF
	t = t << 16 & 0xFFFF0000
	return uint64(t | r)
}

// PUID is a pseudo unique ID, crock32 encoded
func PUID() string {
	p := pseudo(rand.Int63(), time.Now().UnixNano())
	return Encode(p)
}
