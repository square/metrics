package function

import (
	"reflect"
	"testing"

	"github.com/square/metrics/testing_support/assert"
)

func Test_FetchCounter(t *testing.T) {
	c := NewFetchCounter(10)
	a := assert.New(t)
	a.EqInt(c.Current(), 0)
	a.EqInt(c.Limit(), 10)
	a.EqBool(c.Consume(5), true)
	a.EqInt(c.Current(), 5)
	a.EqBool(c.Consume(4), true)
	a.EqInt(c.Current(), 9)
	a.EqBool(c.Consume(1), true)
	a.EqInt(c.Current(), 10)
	a.EqBool(c.Consume(1), false)
	a.EqInt(c.Current(), 11)
}

func TestCopy(t *testing.T) {
	a := EvaluationContext{}
	b := a.Copy()
	if &a == &b {
		t.Errorf("Evaluation context should have been a copy.")
	}
	a.AddNote("Blah")
	if len(b.EvaluationNotes) != 0 {
		t.Errorf("Evaluation context should have been a copy.")
	}
}

func TestNoteCopy(t *testing.T) {
	a := EvaluationContext{}
	a.AddNote("We don't copy notes")
	b := a.Copy()
	if len(b.EvaluationNotes) != 0 {
		t.Errorf("The notes were unexpectedly copied between EvaluationContexts")
	}
	b.AddNote("ABC")
	a.CopyNotesFrom(&b)
	expected := []string{"We don't copy notes", "ABC"}
	if !reflect.DeepEqual(a.EvaluationNotes, expected) {
		t.Errorf("The notes don't match")
	}
}

func TestInvalidation(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected performing an evaluation on an invalid context to panic")
		}
	}()

	ctx := EvaluationContext{}
	ctx.Invalidate()
	EvaluateMany(&ctx, []Expression{})
}
