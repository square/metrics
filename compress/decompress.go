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
