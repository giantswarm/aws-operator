// Copyright 2013 Richard Lehane. All rights reserved.
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

// Package crock32 implements Douglas Crockford's Base32 encoding.
//
// Crock32 is useful for "expressing numbers in a form that can be conveniently and accurately transmitted between humans and computer systems."
// See http://www.crockford.com/wrmg/base32.html for details.
// Note: crock32 differs from Crockford in its use of lower-case letters when encoding (decode works for both cases). To change, use: crock32.SetDigits("0123456789ABCDEFGHJKMNPQRSTVWXYZ")
//
// Example:
//   i, _ := crock32.Decode("a1j3")
//   s := crock32.Encode(i)
//   fmt.Println(s)
package crock32

import "errors"

const cutoff uint64 = (1<<64-1)/32 + 1

// Decode converts a string matching Douglas Crockford's character set (case insensitive) into an unsigned 64-bit integer.
func Decode(s string) (uint64, error) {
	var n uint64
	for i := 0; i < len(s); i++ {
		var v byte
		d := s[i]
		switch {
		case d == 'O', d == 'o':
			v = '0'
		case d == 'L', d == 'l', d == 'I', d == 'i':
			v = '1'
		case '0' <= d && d <= '9':
			v = d - '0'
		case 'a' <= d && d <= 'h':
			v = d - 'a' + 10
		case 'A' <= d && d <= 'H':
			v = d - 'A' + 10
		case 'j' <= d && d <= 'k':
			v = d - 'a' + 9
		case 'J' <= d && d <= 'K':
			v = d - 'A' + 9
		case 'm' <= d && d <= 'n':
			v = d - 'a' + 8
		case 'M' <= d && d <= 'N':
			v = d - 'A' + 8
		case 'p' <= d && d <= 't':
			v = d - 'a' + 7
		case 'P' <= d && d <= 'T':
			v = d - 'A' + 7
		case 'v' <= d && d <= 'z':
			v = d - 'a' + 6
		case 'V' <= d && d <= 'Z':
			v = d - 'A' + 6
		default:
			return 0, errors.New("crock32.Decode: invalid character " + string(d))
		}

		if n >= cutoff {
			return 0, errors.New("crock32.Decode:" + s + " overflows uint64")
		}

		n = n*32 + uint64(v)
	}
	return n, nil
}

var digits = "0123456789abcdefghjkmnpqrstvwxyz"

// SetDigits allows you to change the encoding alphabet (not the decoding alphabet).
// The main purpose of this function is to allow upper-case encoding with crock32.SetDigits("0123456789ABCDEFGHJKMNPQRSTVWXYZ")
func SetDigits(s string) error {
	if len(s) == 32 {
		digits = s
		return nil
	}
	return errors.New("crock32.SetDigits: character set can be anything but it must be 32 characters long")
}

const maxuint = 13

// Encode converts a uint64 into a Crockford base32 encoded string
func Encode(n uint64) string {
	var a [maxuint]byte
	i := maxuint
	for n >= 32 {
		i--
		a[i] = digits[n%32]
		n /= 32
	}
	i--
	a[i] = digits[n]
	return string(a[i:])
}
