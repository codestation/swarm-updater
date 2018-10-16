/*
Copyright 2018 codestation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package log

import "log"

var Printf = log.Printf
var Info = log.Print
var Println = log.Println
var Fatal = log.Fatal
var Fatalf = log.Fatalf
var Debug = emptyFunc

var emptyFunc = func(format string, v ...interface{}) {}

func EnableDebug(debug bool) {
	if debug {
		Debug = log.Printf
	} else {
		Debug = emptyFunc
	}
}
