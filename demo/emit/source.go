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

package main

import (
	"math"
	"math/rand"
)

type Source interface {
	Advance(int)
	Value() float64
}

type Generator struct {
	epoch int
	value float64
	next  func(float64) float64
}

func (g *Generator) Advance(epoch int) {
	if g.epoch == epoch {
		return
	}
	g.epoch = epoch
	g.value = g.next(g.value)
}
func (g *Generator) Value() float64 {
	return g.value
}

func Generate(f func(float64) float64) Source {
	return &Generator{0, 0, f}
}

func Normal() Source {
	return &Generator{0, 0, func(old float64) float64 {
		return rand.NormFloat64()
	}}
}
func Uniform() Source {
	return &Generator{0, 0, func(old float64) float64 {
		return rand.Float64()
	}}
}
func Brownian() Source {
	return &Generator{0, 0, func(old float64) float64 {
		return old + rand.NormFloat64()
	}}
}

type Linear struct {
	sources []Source
	weights []float64
}

func (linear *Linear) Advance(epoch int) {
	for i := range linear.sources {
		linear.sources[i].Advance(epoch)
	}
}
func (linear *Linear) Value() float64 {
	value := 0.0
	for i := range linear.sources {
		value += linear.weights[i] * linear.sources[i].Value()
	}
	return value
}

func NewLinear(sources []Source) Source {
	weights := make([]float64, len(sources))
	weight := 0.0
	for i := range weights {
		weights[i] = rand.Float64()
		weight += weights[i]
	}
	for i := range weights {
		weights[i] /= weight
	}
	return &Linear{sources, weights}
}

type Capper struct {
	source Source
	min    float64
	max    float64
}

func (c *Capper) Advance(epoch int) {
	c.source.Advance(epoch)
}
func (c *Capper) Value() float64 {
	return math.Max(c.min, math.Min(c.max, c.source.Value()))
}

type Cumulative struct {
	source Source
	epoch  int
	sum    float64
}

func (c *Cumulative) Advance(epoch int) {
	c.source.Advance(epoch)
	if c.epoch == epoch {
		return
	}
	c.sum += c.source.Value()
}
func (c *Cumulative) Value() float64 {
	return c.sum
}

type Mapper struct {
	source Source
	fun    func(float64) float64
}

func (m *Mapper) Advance(epoch int) {
	m.source.Advance(epoch)
}
func (m *Mapper) Value() float64 {
	return m.fun(m.source.Value())
}

func RequestSources(count int) []Source {
	base := []Source{Normal(), Uniform(), Brownian(), Brownian(), Brownian()}
	result := make([]Source, count)
	for i := range result {
		result[i] = NewLinear(base)
	}
	return result
}
