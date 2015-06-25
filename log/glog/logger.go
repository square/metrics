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

package glog

import (
	"github.com/golang/glog"
	"github.com/square/metrics/log"
)

type Logger struct{}

func (sl *Logger) Debugf(format string, args ...interface{}) {
	glog.V(1).Infof(format, args...)
}

func (sl *Logger) Infof(format string, args ...interface{}) {
	glog.Infof(format, args...)
}

func (sl *Logger) Warningf(format string, args ...interface{}) {
	glog.Warningf(format, args...)
}

func (sl *Logger) Errorf(format string, args ...interface{}) {
	glog.Errorf(format, args...)
}

func (sl *Logger) Fatalf(format string, args ...interface{}) {
	glog.Fatalf(format, args...)
}

var _ log.Logger = (*Logger)(nil)
