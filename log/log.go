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

// package used to control logging within the square/metrics framework.
package log

var appLogger Logger

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

  // Fatalf only when we can't recover. This should exit after being called.
	Fatalf(format string, args ...interface{})
}

func Debugf(format string, args ...interface{}) {
	if appLogger != nil {
		appLogger.Debugf(format, args)
	}
}

func Infof(format string, args ...interface{}) {
	if appLogger != nil {
		appLogger.Infof(format, args)
	}
}

func Warningf(format string, args ...interface{}) {
	if appLogger != nil {
		appLogger.Warningf(format, args)
	}
}

func Errorf(format string, args ...interface{}) {
  if appLogger != nil {
    appLogger.Errorf(format, args)
  }
}

func Fatalf(format string, args ...interface{}) {
  if appLogger != nil {
    appLogger.Fatalf(format, args)
  }
}

func InitLogger(logger Logger) {
	appLogger = logger
}
