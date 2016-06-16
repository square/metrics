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

package compress

import (
	"bytes"
	"encoding/binary"
	"math"
)

// CompressionBuffer holds compressed data and provides methods to compress it.
type CompressionBuffer struct {
	buffer                  *bytes.Buffer
	current                 uint64
	position                uint32
	leadingZeroWindowLength uint32
	lengthOfWindow          uint32

	finalized   bool
	firstPass   bool
	previousXOR uint64 //The last float we compressed
}

// Bytes returns the remaining bytes from the buffer.
func (c *CompressionBuffer) Bytes() []byte {
	// @@ leaking param: c to result ~r0 level=2
	if !c.finalized {
		panic("Attempted to read bytes from an unfinalized compression buffer")
	}
	return c.buffer.Bytes()
}

// @@ inlining call to Bytes

func (c *CompressionBuffer) fixup() {
	// @@ leaking param content: c
	if c.position == 64 {
		err := binary.Write(c.buffer, binary.BigEndian, c.current)
		if err != nil {
			// @@ c.buffer escapes to heap
			// @@ binary.BigEndian escapes to heap
			// @@ c.current escapes to heap
			panic("WUT")
		}
		c.position = 0
		c.current = 0
	}
}

// NewCompressionBuffer creates a new empty compression buffer.
func NewCompressionBuffer() CompressionBuffer {
	return CompressionBuffer{
		// @@ can inline NewCompressionBuffer
		buffer:   &bytes.Buffer{},
		position: 0,
		// @@ &bytes.Buffer literal escapes to heap
		current:   0,
		firstPass: true,
	}
}

func (c *CompressionBuffer) writeFloat(x float64) {
	// @@ leaking param content: c
	if c.position != 0 {
		panic("Cannot write float when a partial block is waiting on the buffer.")
		// Since other data is flushed to the buffer only every 8 bytes,
		// writing 2 bytes and then a float will actually put those 2 bytes AFTER the float.
		// With this check in place, no confusion should occur.
	}
	err := binary.Write(c.buffer, binary.BigEndian, x)
	if err != nil {
		// @@ c.buffer escapes to heap
		// @@ binary.BigEndian escapes to heap
		// @@ x escapes to heap
		panic("WUT")
	}
}

func (c *CompressionBuffer) writeOne() {
	// @@ leaking param content: c
	c.current = c.current << 1
	c.current = c.current | 1
	c.position++
	c.fixup()
}

func (c *CompressionBuffer) writeZero() {
	// @@ leaking param content: c
	c.current = c.current << 1
	c.position++
	c.fixup()
}

func (c *CompressionBuffer) writeBit(bit bool) {
	// @@ leaking param content: c
	// @@ leaking param content: c
	if bit {
		c.writeOne()
	} else {
		c.writeZero()
	}
}

func (c *CompressionBuffer) writeLowerBits(count uint32, value uint64) {
	// @@ leaking param content: c
	i := count
	for i != MaxUint32 {
		c.writeBit(nthLowestBit(i, value))
		i--
		// @@ inlining call to nthLowestBit
	}
}

func (c *CompressionBuffer) encodeMeaningfulXOR(x uint64) {
	// @@ leaking param content: c
	// @@ leaking param content: c
	// @@ leaking param content: c
	// @@ leaking param content: c
	// @@ leaking param content: c
	leadingZeros := leadingZeros64(x)                   // in the interval [0, 64]
	trailingZeros := trailingZeros64(x)                 // in the interval [0, 64]
	length := uint32(64) - leadingZeros - trailingZeros // in the interval [0, 64]

	meaningfulRegion := x >> trailingZeros

	if leadingZeros == c.leadingZeroWindowLength && length == c.lengthOfWindow && c.lengthOfWindow != 0 {
		//Case A: The meaningful bits of this block have the same # of leading
		//zeros and field size as the previous compressed float.
		c.writeZero()
	} else {
		//Case B: This float has a different # of leading zeros and/or size
		//than the previous one.
		// Describe the number of leading and trailing zeroes.
		c.writeOne()
		c.writeLowerBits(4, uint64(leadingZeros))
		c.writeLowerBits(5, uint64(length))
	}

	// Describe the "meaningful region" of the integer.
	c.writeLowerBits(length, meaningfulRegion)

	c.lengthOfWindow = length
	c.leadingZeroWindowLength = leadingZeros
}

//Finalize the buffer and process whatever
//remaining bytes we have. Omit trailing zero
//valued bytes.
func (c *CompressionBuffer) Finalize() {
	// @@ leaking param content: c
	if c.finalized {
		panic("Cannot finalize a CompressionBuffer twice.")
	}
	for c.position != 0 {
		c.writeZero() // Pad the end with zeros, which flushes the 'current'
	}
	c.finalized = true
}

// Compress takes data as input, compresses it, and puts it into the buffer.
func (c *CompressionBuffer) Compress(data []float64) {
	// @@ leaking param content: c
	// @@ leaking param content: c
	// @@ leaking param content: c
	// @@ leaking param content: c
	for _, value := range data {
		current := math.Float64bits(value)
		if c.firstPass {
			// @@ inlining call to math.Float64bits
			c.writeFloat(value)
			c.previousXOR = current
			c.firstPass = false
			continue
		}
		result := c.previousXOR ^ current
		if result == 0 {
			c.writeZero()
		} else {
			c.writeOne()
			c.encodeMeaningfulXOR(result)
		}

		c.previousXOR = current
	}
}
