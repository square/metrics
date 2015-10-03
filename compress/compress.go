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
	"fmt"
	"math"
)

const MaxUint32 = ^uint32(0)

type CompressionBuffer struct {
	buffer                  *bytes.Buffer
	current                 uint64
	position                uint32
	leadingZeroWindowLength uint64
	lengthOfWindow          uint64

	finalized   bool
	firstPass   bool
	previousXOR uint64 //The last float we compressed
}

type DecompressionBuffer struct {
	data             []byte
	current          byte
	currentByteIndex uint32

	leadingZeroWindowLength uint64
	lengthOfWindow          uint64
	position                uint32
	eof                     bool
	expectedSize            int
}

func NewDecompressionBuffer(data []byte, expectedSize int) DecompressionBuffer {
	dbuf := DecompressionBuffer{
		data:         data,
		position:     uint32(7), // Start reading from the "left"
		eof:          false,
		expectedSize: expectedSize,
	}

	if len(data) <= 8 {
		//Tiny input.
		dbuf.eof = true
	} else {
		dbuf.current = data[8]
		dbuf.currentByteIndex = 8
	}

	return dbuf
}

//The first entry in a new stream is always a completely
//uncompressed 64 bit float.
func (d *DecompressionBuffer) readFirst() float64 {
	b := d.data[0:8]
	buf := bytes.NewReader(b)
	var result float64
	err := binary.Read(buf, binary.BigEndian, &result)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
		panic("Failed to decompress in ReadFirst")
	}

	return result
}

func (d *DecompressionBuffer) hasMore() bool {
	return !d.eof
}

func (d *DecompressionBuffer) ReadBit() uint32 {
	if d.eof {
		panic("Tried reading an invalid bit")
	}

	bit := uint32((d.current & (1 << d.position)) >> d.position)

	d.position--

	if d.position == MaxUint32 {
		if d.currentByteIndex+1 < uint32(len(d.data)) {
			d.position = uint32(7)
			d.currentByteIndex++
			d.current = d.data[d.currentByteIndex]
		} else {
			//No more bytes available.
			d.eof = true
		}
	}
	return bit
}

//The current value we're trying to read has the same number
//of leading zeros and XOR length as the previous entry.
func (d *DecompressionBuffer) readPartialXOR(previous float64) float64 {
	j := uint64(d.lengthOfWindow)
	var xor uint64
	xor = 0
	for uint32(j) != MaxUint32 {
		bit := d.ReadBit()
		xor = xor | (uint64(bit) << j)
		j--
	}

	var rebuiltNumber uint64
	rebuiltNumber = 0
	previousBits := math.Float64bits(previous)
	xor = xor << (64 - d.leadingZeroWindowLength - d.lengthOfWindow)
	rebuiltNumber = previousBits ^ xor

	var buffer *bytes.Buffer
	buffer = new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, rebuiltNumber)

	b := buffer.Bytes()
	readbuf := bytes.NewReader(b)
	var result float64
	err := binary.Read(readbuf, binary.BigEndian, &result)
	if err != nil {
		panic("WUT")
	}

	return result
}

//Read a complete XOR record from the stream. 5 bits for leadering
//zeros, 6 bits for XOR length, and then the XOR field.
func (d *DecompressionBuffer) readFullXOR(previous float64) float64 {
	i := uint32(4)
	var leadingZeros uint32
	leadingZeros = 0
	for i != MaxUint32 {
		bit := d.ReadBit()
		leadingZeros = leadingZeros | (bit << i)
		i--
	}

	i = uint32(5)
	var xorLength uint32
	xorLength = 0
	for i != MaxUint32 {
		bit := d.ReadBit()
		xorLength = xorLength | (bit << i)
		i--
	}

	j := uint64(xorLength)
	var xor uint64
	xor = 0
	for uint32(j) != MaxUint32 {
		bit := d.ReadBit()
		xor = xor | (uint64(bit) << j)
		j--
	}

	var rebuiltNumber uint64
	rebuiltNumber = 0
	previousBits := math.Float64bits(previous)
	xor = xor << (64 - leadingZeros - xorLength)
	rebuiltNumber = previousBits ^ xor

	var buffer *bytes.Buffer
	buffer = new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, rebuiltNumber)

	b := buffer.Bytes()
	readbuf := bytes.NewReader(b)
	var result float64
	err := binary.Read(readbuf, binary.BigEndian, &result)
	if err != nil {
		panic("WUT")
	}

	d.lengthOfWindow = uint64(xorLength)
	d.leadingZeroWindowLength = uint64(leadingZeros)

	return result
}

func (c *CompressionBuffer) Bytes() []byte {
	if !c.finalized {
		panic("Attempted to read bytes from an unfinalized compression buffer")
	}
	return c.buffer.Bytes()
}

func NewCompressionBuffer() CompressionBuffer {
	return CompressionBuffer{
		buffer:    new(bytes.Buffer),
		position:  0,
		current:   0,
		firstPass: true,
	}
}

func (c *CompressionBuffer) writeFirst(x float64) {
	err := binary.Write(c.buffer, binary.BigEndian, x)
	if err != nil {
		panic("WUT")
	}
}

func (c *CompressionBuffer) writeOne() {
	c.current = c.current << uint64(1)
	c.current = c.current | uint64(1)
	c.position++
	c.fixup()
}

func (c *CompressionBuffer) writeZero() {
	c.current = c.current << 1
	c.position++
	c.fixup()
}

func (c *CompressionBuffer) fixup() {
	if c.position == uint32(64) {
		err := binary.Write(c.buffer, binary.BigEndian, c.current)
		if err != nil {
			panic("WUT")
		}
		c.position = 0
		c.current = uint64(0)
	}
}

func (c *CompressionBuffer) encodeMeaningfulXOR(x uint64) {
	leadingZeros := leadingZeros64(x)
	trailingZeros := trailingZeros64(x)
	length := uint64(64) - uint64(leadingZeros) - uint64(trailingZeros)

	c.writeOne()

	//Case A: The meaningful bits of this block have the same # of leading
	//zeros and field size as the previous compressed float.
	if uint64(leadingZeros) == c.leadingZeroWindowLength && uint64(length) == c.lengthOfWindow && c.lengthOfWindow != 0 {
		c.writeZero()
		meaningfulRegion := x >> (64 - c.leadingZeroWindowLength - c.lengthOfWindow)
		c.writeMeaningfulPartialXOR(c.lengthOfWindow, meaningfulRegion)
	} else {
		//Case B: This float has a different # of leading zeros and/or size
		//than the previous one.
		c.writeOne()
		meaningfulRegion := (x >> trailingZeros)
		c.writeMeaningfulXOR(uint64(leadingZeros), length, meaningfulRegion)

		c.lengthOfWindow = length
		c.leadingZeroWindowLength = uint64(leadingZeros)
	}
}

func (c *CompressionBuffer) writeMeaningfulPartialXOR(length uint64, meaningful uint64) {
	i := uint32(length)
	for i != MaxUint32 {
		bit := (meaningful >> i) & 1
		if bit == 1 {
			c.writeOne()
		} else {
			c.writeZero()
		}
		i--
	}
}

func (c *CompressionBuffer) writeMeaningfulXOR(leadingZero uint64, lengthOfMeaning uint64, meaningful uint64) {
	i := uint32(4)
	for i != MaxUint32 {
		bit := (leadingZero >> i) & 1
		if bit == 1 {
			c.writeOne()
		} else {
			c.writeZero()
		}
		i--
	}

	i = uint32(5)
	for i != MaxUint32 {
		bit := (lengthOfMeaning >> i) & 1
		if bit == 1 {
			c.writeOne()
		} else {
			c.writeZero()
		}
		i--
	}

	i = uint32(lengthOfMeaning)
	for i != MaxUint32 {
		bit := (meaningful >> i) & 1
		if bit == 1 {
			c.writeOne()
		} else {
			c.writeZero()
		}
		i--
	}
}

//Finalize the buffer and process whatever
//remaining bytes we have. Omit trailing zero
//valued bytes.
func (c *CompressionBuffer) Finalize() {
	x := c.current
	x = x << (64 - c.position)
	buf := make([]byte, 8)

	binary.BigEndian.PutUint64(buf, x)
	for i, b := range buf {
		//Only write the bytes we need
		if uint32(i*8) <= c.position {
			err := binary.Write(c.buffer, binary.BigEndian, b)
			if err != nil {
				panic("WUT")
			}
		}

	}

	c.finalized = true
}

func leadingZeros64(x uint64) int {
	var upper uint32
	var lower uint32
	lower = (uint32)(x)
	upper = (uint32)(x >> 32)

	zeros := leadingZeros(upper)
	if zeros == 32 {
		zeros += leadingZeros(lower)
	}
	return zeros
}

func leadingZeros(x uint32) int {
	n := 0
	if x == 0 {
		return 32
	}

	if x <= 0x0000FFFF {
		n = n + 16
		x = x << 16
	}
	if x <= 0x00FFFFFF {
		n = n + 8
		x = x << 8
	}
	if x <= 0x0FFFFFFF {
		n = n + 4
		x = x << 4
	}
	if x <= 0x3FFFFFFF {
		n = n + 2
		x = x << 2
	}
	if x <= 0x7FFFFFFF {
		n = n + 1
	}

	return n
}

func trailingZeros64(x uint64) uint32 {
	var upper uint32
	var lower uint32
	lower = (uint32)(x)
	upper = (uint32)(x >> 32)

	zeros := trailingZeros(lower)
	if zeros == 32 {
		zeros += trailingZeros(upper)
	}
	return zeros
}

func trailingZeros(x uint32) uint32 {
	n := uint32(1)
	if x == 0 {
		return 32
	}

	if (x & 0x0000FFFF) == 0 {
		n = n + 16
		x = x >> 16
	}
	if (x & 0x000000FF) == 0 {
		n = n + 8
		x = x >> 8
	}
	if (x & 0x0000000F) == 0 {
		n = n + 4
		x = x >> 4
	}
	if (x & 0x00000003) == 0 {
		n = n + 2
		x = x >> 2
	}
	return n - (x & uint32(1))
}

func (d *DecompressionBuffer) Decompress() []float64 {
	first := d.readFirst()

	result := make([]float64, 1)
	result[0] = first

	var bit uint32
	var prev float64
	prev = first

	for d.hasMore() && len(result) < d.expectedSize {
		bit = d.ReadBit()
		if bit == 0 {
			//Repeat of previous value.
			result = append(result, prev)
		} else {
			//Hit a 1, so we need another bit to know what to do
			bit = d.ReadBit()
			if bit == 1 {
				//Control bit. We have full XOR + lengths.
				num := d.readFullXOR(prev)
				prev = num
				result = append(result, num)
			} else {
				//The next XOR has the same # of leading zeros and length
				//as the previous entry.
				num := d.readPartialXOR(prev)
				prev = num
				result = append(result, num)
			}
		}
	}
	return result
}

func (c *CompressionBuffer) Compress(data []float64) {
	var i int
	if c.firstPass {
		c.writeFirst(data[0])
		c.previousXOR = math.Float64bits(data[0])
		c.firstPass = false
		i = 1
	} else {
		i = 0
	}

	for i < len(data) {
		current := math.Float64bits(data[i])
		result := c.previousXOR ^ current
		if result == 0 {
			c.writeZero()
		} else {
			c.encodeMeaningfulXOR(result)
		}

		c.previousXOR = current
		i++
	}
}
