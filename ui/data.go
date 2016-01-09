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

package ui

import (
	"github.com/square/metrics/inspect"
)

type Config struct {
	Port      int    `yaml:"port"`
	Timeout   int    `yaml:"timeout"`
	StaticDir string `yaml:"static_dir"`
}

type Hook struct {
	OnQuery chan<- *inspect.Profiler
}

type response struct {
	Success  bool                   `json:"success"`
	Name     string                 `json:"name,omitempty"`
	Message  string                 `json:"message,omitempty"`
	Body     interface{}            `json:"body,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Profile  []profileJSON          `json:"profile,omitempty"`
}

type profileJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Start       int64  `json:"start"`  // ms since Unix epoch
	Finish      int64  `json:"finish"` // ms since Unix epoch
}
