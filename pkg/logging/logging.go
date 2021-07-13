// Copyright (c) 2021 Intel Corporation
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

package logging

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"path/filepath"
	"runtime"
)

// Level type
type Level uint32

// Common use of different level:
// "panic": Code crash.
// "error": Unusual event occurred (invalid input or system issue),
//    so exiting code prematurely.
// "warning":  Unusual event occurred (invalid input or system issue),
//    but continuing.
// "info": Basic information, indication of major code paths.
// "debug": Additional information, indication of minor code branches.
//    functions.
const (
	PanicLevel Level = iota
	ErrorLevel
	WarningLevel
	InfoLevel
	DebugLevel
	MaxLevel
	UnknownLevel
)

var loggingStderr bool
var loggingFp *os.File
var loggingLevel Level
var pluginName string

//callDepth sets the number of function calls to retrieve the stack trace for filepath.
const callDepth = 2
const defaultTimestampFormat = time.RFC3339

func (l Level) String() string {
	switch l {
	case PanicLevel:
		return "panic"
	case ErrorLevel:
		return "error"
	case WarningLevel:
		return "warning"
	case InfoLevel:
		return "info"
	case DebugLevel:
		return "debug"
	}
	return "unknown"
}

// Printf prints logging to logfile
func Printf(level Level, format string, a ...interface{}) {
	header := "%s [%s] "
	t := time.Now()
	if level > loggingLevel {
		return
	}

	if loggingLevel == DebugLevel {
		_, path, line, ok := runtime.Caller(callDepth)
		if ok {
			file := filepath.Base(path)
			format = fmt.Sprintf("%s:%d %s", file, line, format)
		}
	}

	if loggingStderr {
		fmt.Fprintf(os.Stderr, header, t.Format(defaultTimestampFormat), level)
		fmt.Fprintf(os.Stderr, format, a...)
		fmt.Fprintf(os.Stderr, "\n")
	}

	if loggingFp != nil {
		fmt.Fprintf(loggingFp, header, t.Format(defaultTimestampFormat), level)
		fmt.Fprintf(loggingFp, format, a...)
		fmt.Fprintf(loggingFp, "\n")
	}
}

// Debugf prints logging if logging level >= debug
func Debugf(format string, a ...interface{}) {
	Printf(DebugLevel, format, a...)
}

// Infof prints logging if logging level >= info
func Infof(format string, a ...interface{}) {
	Printf(InfoLevel, format, a...)
}

// Warningf prints logging if logging level >= Warning
func Warningf(format string, a ...interface{}) {
	Printf(WarningLevel, format, a...)
}

// Errorf prints logging if logging level >= error
func Errorf(format string, a ...interface{}) error {
	Printf(ErrorLevel, format, a...)
	return fmt.Errorf(format, a...)
}

// Panicf prints logging plus stack trace. This should be used only for unrecoverble error
func Panicf(format string, a ...interface{}) {
	Printf(PanicLevel, format, a...)
	Printf(PanicLevel, "========= Stack trace output ========")
	Printf(PanicLevel, "%+v", errors.New("CNDP K8s Plugin Panic"))
	Printf(PanicLevel, "========= Stack trace output end ========")
}

// GetLoggingLevel gets current logging level
func GetLoggingLevel(levelStr string) Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warning":
		return WarningLevel
	case "error":
		return ErrorLevel
	case "panic":
		return PanicLevel
	}
	fmt.Fprintf(os.Stderr, "%s logging: cannot set logging level to %s\n", pluginName, levelStr)
	return UnknownLevel
}

// SetLogLevel sets logging level
func SetLogLevel(levelStr string) {
	level := GetLoggingLevel(levelStr)
	if level < MaxLevel {
		loggingLevel = level
	}
}

// SetLogStderr sets flag for logging stderr output
func SetLogStderr(enable bool) {
	loggingStderr = enable
}

// SetLogFile sets logging file
func SetLogFile(filename string) {
	if filename == "" {
		return
	}

	fp, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		loggingFp = nil
		fmt.Fprintf(os.Stderr, "%s logging: cannot open %s", pluginName, filename)
	}
	loggingFp = fp
}

//SetPluginName sets plugin name
func SetPluginName(PluginStr string) {
	if len(PluginStr) > 0 {
		pluginName = PluginStr
	}
}

func init() {
	loggingStderr = true
	loggingFp = nil
	loggingLevel = WarningLevel
	pluginName = "unnamed plugin"
}