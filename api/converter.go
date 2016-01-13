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

package api

// MetricConverter is an interface that converts metrics from a slice of bytes into a TaggedMetric form.
// Allowing it to act on arbitrary byte sequences gives a client greater freedom in representing their
// metric names.
type MetricConverter interface {
	ToTagged(string) (TaggedMetric, error)
	ToUntagged(TaggedMetric) (string, error)
}

/*
Some notes:
Having an abstract representation for conversion would be a good idea, since in practice it would be helpful
to be able to easily switch to an alternate converter.

But where exactly is it being used? We have TimeseriesStorageAPI for example:
Blueflood needs a GraphiteConverter to know what it's doing with the Tagged metrics.
Maybe it should be given pre-converted values instead?

This would simplify the logic that Blueflood has to do, making its design a LOT more orthogonal to everything else.
*/
