/*
 MIT License

 Copyright (c) 2019 Max Kuznetsov <syhpoon@syhpoon.ca>

 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included in all
 copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 SOFTWARE.
*/

package tattle

import (
	"fmt"
	"time"
)

// Simple Printf Logger implementation
type LoggerPrintf struct {
}

func (log *LoggerPrintf) Debug(format string, args ...interface{}) {
	log.print("DEBUG", format, args...)
}

func (log *LoggerPrintf) Info(format string, args ...interface{}) {
	log.print("INFO", format, args...)
}

func (log *LoggerPrintf) Warning(format string, args ...interface{}) {
	log.print("WARN", format, args...)
}

func (log *LoggerPrintf) Error(format string, args ...interface{}) {
	log.print("ERROR", format, args...)
}

func (log *LoggerPrintf) Critical(format string, args ...interface{}) {
	log.print("CRITICAL", format, args...)
}

func (log *LoggerPrintf) print(level, format string, args ...interface{}) {
	ts := time.Now().Format(time.RFC3339)

	fmt.Printf("%s [%s] %s", ts, level, fmt.Sprintf(format, args))
}
