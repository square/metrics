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

package natural_sort

import (
	"math/rand"
	"testing"

	"github.com/square/metrics/assert"
)

func TestNaturalSort(t *testing.T) {
	a := assert.New(t)
	expected := []string{"Apple", "apple", "file2", "file22", "file90", "file99", "file100", "Zoo", "zoo"}
	tests := [][]string{
		{"Apple", "apple", "file2", "file90", "file99", "file100", "Zoo", "zoo", "file22"},
		{"Zoo", "Apple", "apple", "file100", "file2", "file90", "file99", "zoo", "file22"},
		{"file2", "file90", "apple", "Zoo", "file100", "file22", "file99", "zoo", "Apple"},
	}
	Sort([]string{}) // check that no panic occurs
	for _, test := range tests {
		Sort(test)
		a.Eq(test, expected)
	}
}

func testShuffle(array []string) {
	for len(array) != 0 {
		swapIndex := rand.Intn(len(array))
		array[0], array[swapIndex] = array[swapIndex], array[0]
		array = array[1:]
	}
}

func TestRandom(t *testing.T) {
	a := assert.New(t)
	expected := []string{"Apple", "apple", "file2", "file22", "file90", "file99", "file100", "Zoo", "zoo"}
	test := []string{"Apple", "apple", "file2", "file22", "file90", "file99", "file100", "Zoo", "zoo"}
	Sort(test)
	a.Eq(test, expected)
	for i := 0; i < 1000; i++ {
		testShuffle(test)
		a := a.Contextf("input: %+v", test)
		Sort(test)
		a.Eq(test, expected)
	}
}
