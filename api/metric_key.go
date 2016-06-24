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

package api

// MetricKey is the logical name of a given metric.
// MetricKey should not contain any variable component in it.
type MetricKey string

// MetricKeys is an interface implementing sort.Interface to allow it to be sorted.
type MetricKeys []MetricKey

// Len returns the length of the list of MetricKeys.
func (keys MetricKeys) Len() int {
	return len(keys)
}

// Less tells whether the ith key comes before the jth key.
func (keys MetricKeys) Less(i, j int) bool {
	return keys[i] < keys[j]
}

// Swap swaps the ith and jth keys.
func (keys MetricKeys) Swap(i, j int) {
	keys[i], keys[j] = keys[j], keys[i]
}
