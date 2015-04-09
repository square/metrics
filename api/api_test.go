package api

import (
	"testing"
)

func AssertString(t *testing.T, actual string, expected string) {
	if actual != expected {
		t.Errorf("Expected=[%s], actual=[%s]", expected, actual)
	}
}

func TestTagSet_Serialize(t *testing.T) {
	AssertString(t, NewTagSet().Serialize(), "")
	var ts TagSet = NewTagSet()
	ts["dc"] = "sjc1b"
	ts["env"] = "production"
	AssertString(t, ts.Serialize(), "dc=sjc1b,env=production")
}
