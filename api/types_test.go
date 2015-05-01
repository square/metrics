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
	"testing"

	"github.com/square/metrics/assert"
)

func TestTagSet_Serialize(t *testing.T) {
	a := assert.New(t)
	a.EqString(NewTagSet().Serialize(), "")
	ts := NewTagSet()
	ts["dc"] = "sjc1b"
	ts["env"] = "production"
	a.EqString(ts.Serialize(), "dc=sjc1b,env=production")
}

func TestTagSet_Serialize_Escape(t *testing.T) {
	a := assert.New(t)
	ts := NewTagSet()
	ts["weird=key=1"] = "weird,value"
	ts["weird=key=2"] = "weird\\value"
	a.EqString(ts.Serialize(), "weird\\=key\\=1=weird\\,value,weird\\=key\\=2=weird\\\\value")
	parsed := ParseTagSet(ts.Serialize())
	a.EqInt(len(parsed), 2)
	a.EqString(parsed["weird=key=1"], "weird,value")
	a.EqString(parsed["weird=key=2"], "weird\\value")
}

func TestTagSet_ParseTagSet(t *testing.T) {
	a := assert.New(t)
	a.EqString(ParseTagSet("foo=bar").Serialize(), "foo=bar")
	a.EqString(ParseTagSet("a=1,b=2").Serialize(), "a=1,b=2")
	a.EqString(ParseTagSet("a\\,b=1").Serialize(), "a\\,b=1")
	a.EqString(ParseTagSet("a\\=b=1").Serialize(), "a\\=b=1")
}
