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
func (d *DecompressionBuffer) readFloat() float64 {
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

func (d *DecompressionBuffer) ReadBit() bool {
	if d.eof {
		panic("Tried reading an invalid bit")
	}

	bit := nthLowestBit(d.position, uint64(d.current))

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
	return math.Float64frombits(previousBits ^ xor)
}

//Read a complete XOR record from the stream. 5 bits for leadering
//zeros, 6 bits for XOR length, and then the XOR field.
func (d *DecompressionBuffer) readFullXOR(previous float64) float64 {
	leadingZeros := uint32(d.ReadBits(4))
	xorLength := uint32(d.ReadBits(5))

	xor := d.ReadBits(xorLength) << (64 - leadingZeros - xorLength)

	rebuiltNumber := math.Float64bits(previous) ^ xor

	d.lengthOfWindow = xorLength
	d.leadingZeroWindowLength = leadingZeros

	return math.Float64frombits(rebuiltNumber)
}

func (d *DecompressionBuffer) Decompress() []float64 {
	first := d.readFloat()
	result := []float64{first}

	number := first
	for d.hasMore() && len(result) < d.expectedSize {
		if d.ReadBit() {
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
