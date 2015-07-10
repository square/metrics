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
	"sort"
	"unicode"
)

// naturalStringLess returns true when `a` is less than `b`:
// "apples" < "Apples" < "cats1" < "cats2" < "cats10" < "cats20" < "cats100" < "dogs"
func Less(strA string, strB string) bool {
	runesA, runesB := []rune(strA), []rune(strB)
	iA, iB := 0, 0
	for iA < len(runesA) && iB < len(runesB) {
		runeA, runeB := runesA[iA], runesB[iB]
		if unicode.IsDigit(runeA) != unicode.IsDigit(runeB) {
			return unicode.IsDigit(runeA)
		}
		if unicode.IsDigit(runeA) {
			// Then both are digits
			eA, eB := iA, iB
			for eA < len(runesA) && unicode.IsDigit(runesA[eA]) {
				eA++
			}
			for eB < len(runesB) && unicode.IsDigit(runesB[eB]) {
				eB++
			}
			digitsA, digitsB := runesA[iA:eA], runesB[iB:eB]
			// Check if A is shorter (and therefore smaller) than B
			if len(digitsA) != len(digitsB) {
				return len(digitsA) < len(digitsB)
			}
			// Otherwise, just use regular string comparisons (if they're different)
			if string(digitsA) != string(digitsB) {
				return string(digitsA) < string(digitsB)
			}
			iA, iB = eA, eB
			continue
		}
		if runeA != runeB {
			// Compare their lowercase versions (so that A < a < B < b; instead of ASCII A < B < a < b)
			return unicode.ToLower(runeA) < unicode.ToLower(runeB)
		}
	}
	return len(runesA) < len(runesB)
}

type naturalStrings []string

func (a naturalStrings) Len() int {
	return len([]string(a))
}

func (a naturalStrings) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a naturalStrings) Less(i, j int) bool {
	return Less(a[i], a[j])
}

func Sort(array []string) {
	sort.Sort(naturalStrings(array))
}
