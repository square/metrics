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

package inspect

import "time"

// Profiler contains a sequence of profiles which are collected over the course of a query execution.
type Profiler struct {
	profiles []Profile
}

// Add will create a profile of the given name from `start` until the current time.
// Add acts in a threadsafe manner.
func (p *Profiler) Add(name string) func() {
	if p == nil {
		// If the profiler instance doesn't exist, then don't attempt to operate on it.
		return func() {}
	}
	start := time.Now()
	return func() {
		p.profiles = append(p.profiles, Profile{name: name, startTime: start, finishTime: time.Now()})
	}
}

// All returns the list of profiles collected thus far.
func (p *Profiler) All() []Profile {
	if p == nil {
		// If the profiler instance doesn't exist, then don't attempt to operate on it.
		return []Profile{}
	}
	return p.profiles
}

// A profile is a single data point collected by the profiler.
type Profile struct {
	name       string    // name identifies the measured quantity ("fetchSingle() or api.GetAllMetrics()")
	startTime  time.Time // the start time of the task
	finishTime time.Time // the end time of the task
}

// Name is the name of the profile.
func (p Profile) Name() string {
	return p.name
}

// Start is the start time of the profile.
func (p Profile) Start() time.Time {
	return p.startTime
}

// Finish is the finish time of the profile.
func (p Profile) Finish() time.Time {
	return p.finishTime
}

// Duration is the duration of the profile (Finish - Start).
func (p Profile) Duration() time.Duration {
	return p.finishTime.Sub(p.startTime)
}
