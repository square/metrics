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

import (
	"bytes"
)

// Predicate is a boolean function applied against the given
// metric alias and tagset. It determines whether the given metric
// should be included in the query.
type Predicate interface {
	// Since all Predicates satisfy the query.Node interface anyways,
	// embed it here.
	Print(buf *bytes.Buffer, indent int)
	// checks the matcher.
	Apply(tagSet TagSet) bool
}
