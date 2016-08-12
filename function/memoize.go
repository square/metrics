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

package function

import "sync"

// memoized is a synchronized container for the results of an evaluation.
// In order to use it, acquire the lock and then check whether "done" is true.
type memoized struct {
	sync.Mutex
	done  bool
	value Value
	err   error
}

// compute uses the given expression and context to assign the value and err of
// the memoized object, unless they've already been set, or are currently being
// set, in which case it waits for them to complete and then returns the same
// value without re-computing.
func (m *memoized) compute(e ActualExpression, context EvaluationContext) (Value, error) {
	m.Lock()
	defer m.Unlock()
	if m.done {
		return m.value, m.err
	}
	m.value, m.err = e.ActualEvaluate(context)
	m.done = true
	return m.value, m.err
}

type memoization struct {
	sync.Mutex
	memoized map[string]*memoized
}

// evaluate uses the expression's StringMemoization to look it up in the internal
// map. If
func (m *memoization) evaluate(e ActualExpression, context EvaluationContext) (Value, error) {
	if m == nil || m.memoized == nil {
		// if uninitialized, it will always compute the given expressions.
		return e.ActualEvaluate(context)
	}
	m.Lock()
	memoIdentity := e.ExpressionDescription(StringMemoization)
	ptr, ok := m.memoized[memoIdentity]
	if !ok {
		ptr = new(memoized)
		m.memoized[memoIdentity] = ptr
	}
	m.Unlock()
	return ptr.compute(e, context)
}

func newMemo() *memoization {
	return &memoization{
		memoized: make(map[string]*memoized),
	}
}

// A memoizedExpression wraps an expression such that Evaluate is automatically
// memoized.
type memoizedExpression struct {
	Expression ActualExpression
}

// Memoize takes an ordinary actual expression and turns it into a memoized expression.
func Memoize(expression ActualExpression) Expression {
	return memoizedExpression{Expression: expression}
}

// Literal exposes the underlying Expression's literal
func (m memoizedExpression) Literal() interface{} {
	literalExpression, ok := m.Expression.(LiteralExpression)
	if !ok {
		return nil
	}
	return literalExpression.Literal()
}

// Evaluate calls EvaluateMemoized on the underlying expression.
func (m memoizedExpression) Evaluate(context EvaluationContext) (Value, error) {
	return context.EvaluateMemoized(m.Expression)
}

// ExpressionDescription behaves identically to the underlying expression
func (m memoizedExpression) ExpressionDescription(mode DescriptionMode) string {
	return m.Expression.ExpressionDescription(mode)
}

// memoization map holds a collection of memoization points.
type memoizationMap struct {
	sync.Mutex
	Map map[contextIdentity]*memoization
}

func (m *memoizationMap) get(i contextIdentity) *memoization {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.Map[i]; !ok {
		m.Map[i] = newMemo()
	}
	return m.Map[i]
}

func newMemoMap() *memoizationMap {
	return &memoizationMap{Map: map[contextIdentity]*memoization{}}
}
