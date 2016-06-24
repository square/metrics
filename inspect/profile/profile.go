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

package profile

import (
	"sync"
	"time"
)

// Profiler contains a sequence of profiles which are collected over the course of a query execution.
type Profiler struct {
	now      func() time.Time
	mutex    sync.Mutex // Since profilers are only ever used as pointers, the mutex is not a pointer.
	profiles []Profile
}

func New() *Profiler {
	return &Profiler{
		now:      time.Now,
		mutex:    sync.Mutex{},
		profiles: []Profile{},
	}
}

func (p *Profiler) AddProfile(profile Profile) {
	if p == nil {
		return
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.profiles = append(p.profiles, profile)
}

// Record will create a profile of the given name from `start` until the current time.
// Record acts in a threadsafe manner.
func (p *Profiler) Record(name string) func() {
	if p == nil {
		// If the profiler instance doesn't exist, then don't attempt to operate on it.
		return func() {}
	}
	start := p.now()
	return func() {
		p.AddProfile(Profile{
			Name:   name,
			Start:  start,
			Finish: p.now(),
		})
	}
}

// Do will perform and time the action given.
// It behaves in a threadsafe manner.
// If the profiler is nil, the action will be performed, but no profile will be recorded.
func (p *Profiler) Do(name string, action func()) {
	if p == nil {
		// If the profiler instance doesn't exist, then don't attempt to operate on it.
		// Make sure that you still run the action
		action()
		return
	}
	start := p.now()
	action()
	p.AddProfile(Profile{
		Name:   name,
		Start:  start,
		Finish: p.now(),
	})
}

func (p *Profiler) RecordWithDescription(name string, description string) func() {
	if p == nil {
		// If the profiler instance doesn't exist, then don't attempt to operate on it.
		return func() {}
	}
	start := p.now()
	return func() {
		p.AddProfile(Profile{
			Name:        name,
			Description: description,
			Start:       start,
			Finish:      p.now(),
		})
	}
}

// All retrieves all the profiling information collected by the profiler.
func (p *Profiler) All() []Profile {
	if p == nil {
		// If the profiler instance doesn't exist, then don't attempt to operate on it.
		return []Profile{}
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.profiles
}

// Flush provides a safe way to clear the profiles from its list.
// It's guaranteed that no profiles will be lost by calling this method.
func (p *Profiler) Flush() []Profile {
	if p == nil {
		return []Profile{}
	}
	p.mutex.Lock()
	defer p.mutex.Unlock()
	result := p.profiles
	p.profiles = []Profile{}
	return result
}

// A Profile is a single data point collected by the profiler.
type Profile struct {
	Name        string    `json:"name"` // name identifies the measured quantity ("fetchSingle() or api.GetAllMetrics()")
	Description string    `json:"description,omitempty"`
	Start       time.Time `json:"start"`  // the start time of the task
	Finish      time.Time `json:"finish"` // the end time of the task
}

// Duration is the duration of the profile (Finish - Start).
func (p Profile) Duration() time.Duration {
	return p.Finish.Sub(p.Start)
}
