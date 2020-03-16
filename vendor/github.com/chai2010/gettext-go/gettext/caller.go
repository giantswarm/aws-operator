// Copyright 2013 ChaiShushan <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gettext

import (
	"regexp"
	"runtime"
)

var (
	reInit1 = regexp.MustCompile(`init·\d+$`)  // main.init·1
	reInit2 = regexp.MustCompile(`init\.\d+$`) // main.init.1

	reClosure1 = regexp.MustCompile(`func·\d+$`)                // main.func·001
	reClosure2 = regexp.MustCompile(`glob\.\.func\d+(\.\d+)*$`) // main.glob..func1
	reClosure3 = regexp.MustCompile(`\w+\.func\d+(\.\d+)*$`)    // main.FuncName.func1
)

// caller types:
// runtime.goexit
// runtime.main
// main.init
// main.main
// main.init·1 -> main.init
// main.func·001 -> main.func
// code.google.com/p/gettext-go/gettext.TestCallerName
// ...
func callerName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	name := runtime.FuncForPC(pc).Name()

	if reInit1.MatchString(name) {
		return reInit1.ReplaceAllString(name, "init")
	}
	if reInit2.MatchString(name) {
		return reInit2.ReplaceAllString(name, "init")
	}

	if reClosure1.MatchString(name) {
		return reClosure1.ReplaceAllString(name, "func")
	}
	if reClosure2.MatchString(name) {
		return reClosure2.ReplaceAllString(name, "func")
	}
	if reClosure3.MatchString(name) {
		return regexp.MustCompile(`func\d+(\.\d+)?$`).ReplaceAllString(name, "func") //+ "##"
	}

	return name
}
