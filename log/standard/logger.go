// Copyright 2015 Square Inc.
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

type StandardLogger struct {
  Logger *standard_logger.Logger
}

func (sl *StandardLogger) logf(level string, format string, args ...interface{}) {
  sl.Logger.Printf(level + " " + format, args)
}

func (sl *StandardLogger) Debugf(format string, args ...interface{}) {
  sl.logf("DEBUG", format, args)
}

func (sl *StandardLogger) Infof(format string, args ...interface{}) {
  sl.logf("INFO", format, args)
}

func (sl *StandardLogger) Warningf(format string, args ...interface{}) {
  sl.logf("WARNING", format, args)
}

func (sl *StandardLogger) Errorf(format string, args ...interface{}) {
  sl.logf("ERROR", format, args)
}

func (sl *StandardLogger) Fatalf(format string, args ...interface{}) {
  sl.Logger.Fatalf(format, args)
}

var _ log.Logger = (*StandardLogger)(nil)
