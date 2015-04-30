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
