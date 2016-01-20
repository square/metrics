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

import "sync"

// EvaluationNotes stores a list of notes from the evaluation.
type EvaluationNotes struct {
	mutex sync.Mutex
	notes []string
}

func (notes *EvaluationNotes) AddNote(note string) {
	if notes == nil {
		return
	}
	notes.mutex.Lock()
	defer notes.mutex.Unlock()
	notes.notes = append(notes.notes, note)
}
func (notes *EvaluationNotes) Notes() []string {
	if notes == nil {
		return nil
	}
	notes.mutex.Lock()
	defer notes.mutex.Unlock()
	return notes.notes
}
