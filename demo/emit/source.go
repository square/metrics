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

package main

import (
	"math"
	"math/rand"
)

// The Source interface represents a source of randomized data.
type Source interface {
	Advance(int)    // Advance sets the source's time.
	Value() float64 // Value gets the source's current value
}

// Generator is a generic type implemening Source that's based on a randomized
// "next" function mapping float64 -> float64.
type Generator struct {
	epoch int
	value float64
	next  func(float64) float64
}

// Advance will update the value stored in the generator if the epoch is new.
func (g *Generator) Advance(epoch int) {
	if g.epoch == epoch {
		return
	}
	g.epoch = epoch
	g.value = g.next(g.value)
}

// Value will return the correct value stored in the generator.
func (g *Generator) Value() float64 {
	return g.value
	// @@ can inline (*Generator).Value
}

// Generate takes a successor function and creates a source.
func Generate(f func(float64) float64) Source {
	// @@ leaking param: f to result ~r1 level=-1
	return &Generator{0, 0, f}
	// @@ can inline Generate
}

// @@ &Generator literal escapes to heap
// @@ &Generator literal escapes to heap

// Normal is a source of normally distributed data points.
func Normal() Source {
	return &Generator{0, 0, func(old float64) float64 {
		return rand.NormFloat64()
		// @@ &Generator literal escapes to heap
		// @@ &Generator literal escapes to heap
		// @@ func literal escapes to heap
		// @@ func literal escapes to heap
	}}
}

// Uniform is a source of uniformly distributed data points on [0, 1].
func Uniform() Source {
	return &Generator{0, 0, func(old float64) float64 {
		return rand.Float64()
		// @@ &Generator literal escapes to heap
		// @@ &Generator literal escapes to heap
		// @@ func literal escapes to heap
		// @@ func literal escapes to heap
	}}
}

// Brownian is a source of brownian data points (first differences are normal).
func Brownian() Source {
	return &Generator{0, 0, func(old float64) float64 {
		return old + rand.NormFloat64()
		// @@ &Generator literal escapes to heap
		// @@ &Generator literal escapes to heap
		// @@ func literal escapes to heap
		// @@ func literal escapes to heap
	}}
}

// Linear computes a linear mix of sources.
type Linear struct {
	sources []Source
	weights []float64
}

// Advance advances every source in the linear collection.
func (linear *Linear) Advance(epoch int) {
	// @@ leaking param content: linear
	for i := range linear.sources {
		linear.sources[i].Advance(epoch)
	}
}

// Value gets the linear mix of the value's of all the sources in the linear container.
func (linear *Linear) Value() float64 {
	// @@ leaking param content: linear
	value := 0.0
	for i := range linear.sources {
		value += linear.weights[i] * linear.sources[i].Value()
	}
	return value
}

// NewLinear takes a slice of sources as input and assigns weights randomly to each.
func NewLinear(sources []Source) Source {
	// @@ leaking param: sources to result ~r1 level=-1
	weights := make([]float64, len(sources))
	weight := 0.0
	// @@ make([]float64, len(sources)) escapes to heap
	// @@ make([]float64, len(sources)) escapes to heap
	for i := range weights {
		weights[i] = rand.Float64()
		weight += weights[i]
	}
	for i := range weights {
		weights[i] /= weight
	}
	return &Linear{sources, weights}
}

// @@ &Linear literal escapes to heap
// @@ &Linear literal escapes to heap

// Capper is an implementation for Source that caps its outputs (above and below).
type Capper struct {
	source Source
	min    float64
	max    float64
}

// Advance advances the underlying source.
func (c *Capper) Advance(epoch int) {
	// @@ leaking param content: c
	c.source.Advance(epoch)
}

// Value gets the capped source for the capper.
func (c *Capper) Value() float64 {
	// @@ leaking param content: c
	return math.Max(c.min, math.Min(c.max, c.source.Value()))
}

// Cumulative is an implementation for Source that computes a running sum of outputs.
type Cumulative struct {
	source Source
	epoch  int
	sum    float64
}

// Advance advances the underlying source.
func (c *Cumulative) Advance(epoch int) {
	// @@ leaking param content: c
	// @@ leaking param content: c
	c.source.Advance(epoch)
	if c.epoch == epoch {
		return
	}
	c.sum += c.source.Value()
}

// Value gets the current value in the sum.
func (c *Cumulative) Value() float64 {
	return c.sum
	// @@ can inline (*Cumulative).Value
}

// Mapper changes the distribution of the underlying source by remapping all outputs.
type Mapper struct {
	source Source
	fun    func(float64) float64
}

// Advance advances the underlying container.
func (m *Mapper) Advance(epoch int) {
	// @@ leaking param content: m
	m.source.Advance(epoch)
}

// Value computes the redistributed value of the underlying source.
func (m *Mapper) Value() float64 {
	// @@ leaking param content: m
	return m.fun(m.source.Value())
}

// RequestSources provides a random collection of several sources.
func RequestSources(count int) []Source {
	base := []Source{Normal(), Uniform(), Brownian(), Brownian(), Brownian()}
	result := make([]Source, count)
	// @@ []Source literal escapes to heap
	for i := range result {
		// @@ make([]Source, count) escapes to heap
		// @@ make([]Source, count) escapes to heap
		result[i] = NewLinear(base)
	}
	return result
}
