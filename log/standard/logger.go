// Copyright 2015 - 2016 Square Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package standard

import (
	standard_logger "log"

	"github.com/square/metrics/log"
)

type Logger struct {
	Logger *standard_logger.Logger
}

func (sl *Logger) logf(format string, level string, args ...interface{}) {
	sl.Logger.Printf(level+" "+format, args...)
}

func (sl *Logger) Debugf(format string, args ...interface{}) {
	sl.logf(format, "DEBUG", args...)
}

func (sl *Logger) Infof(format string, args ...interface{}) {
	sl.logf(format, "INFO", args...)
}

func (sl *Logger) Warningf(format string, args ...interface{}) {
	sl.logf(format, "WARNING", args...)
}

func (sl *Logger) Errorf(format string, args ...interface{}) {
	sl.logf(format, "ERROR", args...)
}

func (sl *Logger) Fatalf(format string, args ...interface{}) {
	sl.Logger.Fatalf(format, args...)
}

var _ log.Logger = (*Logger)(nil)
