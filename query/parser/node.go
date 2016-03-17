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

package parser

import "github.com/square/metrics/timeseries"

type any interface{} // fixes a bug in gopeg

// temporary nodes
// ---------------
// These nodes are only present during the parsing step and are not present
// in the resulting command.
// There are three types of temporary nodes:
// * literals (constants in the syntax tree).
// * lists
// * evaluation context nodes

type stringLiteral struct {
	literal string
}

// list of literals
type stringLiteralList struct {
	literals []string
}

// single tag
type tagLiteral struct {
	tag string
}

// a single operator
type operatorLiteral struct {
	operator string
}

type groupByList struct {
	list      []string
	collapses bool
}

// evaluationContextKey represents a key (from, to, sampleby) for the evaluation context.
type evaluationContextKey struct {
	key string
}

// evaluationContextValue represents a value (date, samplingmode, etc.) for the evaluation context.
type evaluationContextValue struct {
	value string
}

// evaluationContextMap represents a collection of key-value pairs that form the evaluation context.
type evaluationContextNode struct {
	Start        int64                   // Start of data timerange
	End          int64                   // End of data timerange
	Resolution   int64                   // Resolution of data timerange
	SampleMethod timeseries.SampleMethod // to use when up/downsampling to match requested resolution
	assigned     map[string]bool         // a map for knowing which elements of the context have been assigned
}
