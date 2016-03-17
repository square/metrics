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

package compress

import (
	"bytes"
	"encoding/binary"
	"math"
)

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

func (c *CompressionBuffer) Bytes() []byte {
	if !c.finalized {
		panic("Attempted to read bytes from an unfinalized compression buffer")
	}
	return c.buffer.Bytes()
}

func (c *CompressionBuffer) fixup() {
	if c.position == 64 {
		err := binary.Write(c.buffer, binary.BigEndian, c.current)
		if err != nil {
			panic("WUT")
		}
		c.position = 0
		c.current = 0
	}
}

func NewCompressionBuffer() CompressionBuffer {
	return CompressionBuffer{
		buffer:    &bytes.Buffer{},
		position:  0,
		current:   0,
		firstPass: true,
	}
}

func (c *CompressionBuffer) writeFloat(x float64) {
	if c.position != 0 {
		panic("Cannot write float when a partial block is waiting on the buffer.")
		// Since other data is flushed to the buffer only every 8 bytes,
		// writing 2 bytes and then a float will actually put those 2 bytes AFTER the float.
		// With this check in place, no confusion should occur.
	}
	err := binary.Write(c.buffer, binary.BigEndian, x)
	if err != nil {
		panic("WUT")
	}
}

func (c *CompressionBuffer) writeOne() {
	c.current = c.current << 1
	c.current = c.current | 1
	c.position++
	c.fixup()
}

func (c *CompressionBuffer) writeZero() {
	c.current = c.current << 1
	c.position++
	c.fixup()
}

func (c *CompressionBuffer) writeBit(bit bool) {
	if bit {
		c.writeOne()
	} else {
		c.writeZero()
	}
}

func (c *CompressionBuffer) writeLowerBits(count uint32, value uint64) {
	i := count
	for i != MaxUint32 {
		c.writeBit(nthLowestBit(i, value))
		i--
	}
}

func (c *CompressionBuffer) encodeMeaningfulXOR(x uint64) {
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
	if c.finalized {
		panic("Cannot finalize a CompressionBuffer twice.")
	}
	for c.position != 0 {
		c.writeZero() // Pad the end with zeros, which flushes the 'current'
	}
	c.finalized = true
}

func (c *CompressionBuffer) Compress(data []float64) {
	for _, value := range data {
		current := math.Float64bits(value)
		if c.firstPass {
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
