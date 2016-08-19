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
	_ "fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestCompressionRoundtrip(t *testing.T) {
	data := []float64{1.0, 1.3, 1.4, 1.5, 1.6, 2.0, 2.1, 1.1, 1.2, 1.2, 1.2, 0.4, 6.00065e+06, 6.000656e+06, 6.000657e+06, 6.000659e+06, 6.000661e+06, 1.79e+308, 1.79e+308, 1.79e-308, 1.79e-308}
	c := NewCompressionBuffer()
	c.Compress(data)
	c.Finalize()
	compressed := c.Bytes()
	dbuf := NewDecompressionBuffer(compressed, len(data))
	decompressed := dbuf.Decompress()
	if !reflect.DeepEqual(data, decompressed) {
		t.Errorf("The array didn't decompress correctly:\n\tinput:  %v\n\toutput: %v", data, decompressed)
		for i := range data {
			if !reflect.DeepEqual(data[i], decompressed[i]) {
				t.Errorf("\tdata[%d] = %f != decompressed[%d] = %f", i, data[i], i, decompressed[i])
			}
		}
	}
}

func TestMultipleSequentialInputs(t *testing.T) {
	data1 := []float64{1.111, 1.222, 1.333}
	data2 := []float64{1.444, 1.555, 1.666}
	c := NewCompressionBuffer()
	c.Compress(data1)
	c.Compress(data2)
	c.Finalize()
	compressed := c.Bytes()
	dbuf := NewDecompressionBuffer(compressed, len(data1)+len(data2))
	decompressed := dbuf.Decompress()
	expected := append(data1, data2...)
	if !reflect.DeepEqual(expected, decompressed) {
		t.Errorf("The joined array is different.\n%f\n%f\n", expected, decompressed)
	}
}

func TestSmallInput(t *testing.T) {
	data := []float64{1.0}

	c := NewCompressionBuffer()
	c.Compress(data)
	c.Finalize()
	compressed := c.Bytes()
	dbuf := NewDecompressionBuffer(compressed, len(data))
	decompressed := dbuf.Decompress()
	if !reflect.DeepEqual(data, decompressed) {
		t.Errorf("The array didn't decompress correctly.")
	}
}

func TestCompressionLarge(t *testing.T) {
	r := rand.New(rand.NewSource(800))
	length := 100000
	data := make([]float64, length)
	for i := 0; i < length; i++ {
		//To be fair, this really highlights the worst case.
		data[i] = r.ExpFloat64()
	}

	c := NewCompressionBuffer()
	c.Compress(data)
	c.Finalize()
	compressed := c.Bytes()

	dbuf := NewDecompressionBuffer(compressed, len(data))
	decompressed := dbuf.Decompress()
	if !reflect.DeepEqual(data, decompressed) {
		t.Errorf("The array didn't decompress correctly.\n%f\n%f\n", data, decompressed)
	}
}

func TestCompressionRatio(t *testing.T) {
	r := rand.New(rand.NewSource(0))
	mean := 0.0
	count := 100
	for i := 0; i < count; i++ {
		length := r.Intn(5000) + 100
		c := NewCompressionBuffer()
		data := []float64{}
		value := rand.ExpFloat64() * 1e-5
		for j := 0; j < length; j++ {
			if rand.Intn(10) != 0 {
				value += rand.ExpFloat64()
			}
			data = append(data, value)
		}
		c.Compress(data)
		c.Finalize()
		compressed := c.Bytes()
		compressionRatio := float64(length*8) / float64(len(compressed))
		if compressionRatio < 1 {
			t.Errorf("Data size was increased when compressed")
			t.Errorf("Compression ratio: %f; original %d: compressed %d", float64(length*8)/float64(len(compressed)), length*8, len(compressed))
		}
		mean += compressionRatio
	}
	mean = mean / float64(count)
	t.Logf("mean compression ratio: %f", mean)
}
