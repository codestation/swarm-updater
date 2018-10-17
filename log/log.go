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

// Println calls Output to print to the standard logger.
var Println = log.Println
// Fatal is equivalent to Print() followed by a call to os.Exit(1).
var Fatal = log.Fatal
// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
var Fatalf = log.Fatalf
// Printf calls Output to print to the standard logger.
var Printf = log.Printf
// Debug calls Output to print to the standard logger if enabled.
var Debug = nullLogger

func nullLogger(_ string, _ ...interface{}) {
	// no-op
}

// EnableDebug show/hides debug logs
func EnableDebug(debug bool) {
	if debug {
		Debug = log.Printf
	} else {
		Debug = nullLogger
	}
}
