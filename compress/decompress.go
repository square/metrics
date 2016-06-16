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
	"fmt"
	"math"
)

// DecompressionBuffer holds compressed data and provides methods for decompressing it.
type DecompressionBuffer struct {
	data             []byte
	current          byte
	currentByteIndex uint32

	leadingZeroWindowLength uint32
	lengthOfWindow          uint32
	position                uint32
	eof                     bool
	expectedSize            int
}

// NewDecompressionBuffer creates a buffer with the given compressed contents.
func NewDecompressionBuffer(data []byte, expectedSize int) DecompressionBuffer {
	// @@ leaking param: data to result ~r2 level=0
	dbuf := DecompressionBuffer{
		// @@ can inline NewDecompressionBuffer
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
func (d *DecompressionBuffer) readFloat() float64 {
	// @@ leaking param content: d
	b := d.data[0:8]
	buf := bytes.NewReader(b)
	var result float64
	// @@ inlining call to bytes.NewReader
	// @@ &bytes.Reader literal escapes to heap
	err := binary.Read(buf, binary.BigEndian, &result)
	// @@ moved to heap: result
	if err != nil {
		// @@ buf escapes to heap
		// @@ binary.BigEndian escapes to heap
		// @@ &result escapes to heap
		// @@ &result escapes to heap
		fmt.Println("binary.Read failed:", err)
		panic("Failed to decompress in ReadFirst")
		// @@ "binary.Read failed:" escapes to heap
		// @@ err escapes to heap
	}

	return result
}

func (d *DecompressionBuffer) hasMore() bool {
	return !d.eof
	// @@ can inline (*DecompressionBuffer).hasMore
}

// ReadBit returns a single bit from the buffer.
func (d *DecompressionBuffer) ReadBit() bool {
	if d.eof {
		panic("Tried reading an invalid bit")
	}

	bit := nthLowestBit(d.position, uint64(d.current))

	// @@ inlining call to nthLowestBit
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

// ReadBits returns several bits (up to 64) from the buffer in a uint64.
func (d *DecompressionBuffer) ReadBits(n uint32) uint64 {
	value := uint64(0)
	for n != MaxUint32 {
		value = value << 1
		if d.ReadBit() {
			value |= 1
		}
		n--
	}
	return value
}

//The current value we're trying to read has the same number
//of leading zeros and XOR length as the previous entry.
func (d *DecompressionBuffer) readPartialXOR(previous float64) float64 {
	previousBits := math.Float64bits(previous)
	xor := d.ReadBits(d.lengthOfWindow) << (64 - d.leadingZeroWindowLength - d.lengthOfWindow)
	// @@ inlining call to math.Float64bits
	return math.Float64frombits(previousBits ^ xor)
}

// @@ inlining call to math.Float64frombits

//Read a complete XOR record from the stream. 5 bits for leadering
//zeros, 6 bits for XOR length, and then the XOR field.
func (d *DecompressionBuffer) readFullXOR(previous float64) float64 {
	leadingZeros := uint32(d.ReadBits(4))
	xorLength := uint32(d.ReadBits(5))

	xor := d.ReadBits(xorLength) << (64 - leadingZeros - xorLength)

	rebuiltNumber := math.Float64bits(previous) ^ xor

	// @@ inlining call to math.Float64bits
	d.lengthOfWindow = xorLength
	d.leadingZeroWindowLength = leadingZeros

	return math.Float64frombits(rebuiltNumber)
}

// @@ inlining call to math.Float64frombits

// Decompress uses the compressed buffer contents to create a []float64.
func (d *DecompressionBuffer) Decompress() []float64 {
	// @@ leaking param content: d
	first := d.readFloat()
	result := []float64{first}

	// @@ []float64 literal escapes to heap
	number := first
	for d.hasMore() && len(result) < d.expectedSize {
		if d.ReadBit() {
			// @@ inlining call to (*DecompressionBuffer).hasMore
			// Hit a 1, so we need another bit to know what to do.
			// Otherwise it's a repeat of the previous value.
			if d.ReadBit() {
				// With have full XOR + lengths
				number = d.readFullXOR(number)
			} else {
				// We have partial XOR (it has the same number of leading zeroes and length)
				number = d.readPartialXOR(number)
			}
		}
		result = append(result, number)
	}
	return result
}
