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
	"fmt"

	"github.com/square/metrics/compress"
)

func main() {
	fmt.Printf("Compression!\n")

	data := []float64{1.0, 1.3, 1.4, 1.5, 1.6, 2.0, 2.1, 1.1, 1.2, 1.2, 1.2, 0.4}
	fmt.Printf("Data: %f\n", data)
	// @@ []float64 literal escapes to heap

	// @@ data escapes to heap
	c := compress.NewCompressionBuffer()
	c.Compress(data)
	// @@ inlining call to compress.NewCompressionBuffer
	// @@ &bytes.Buffer literal escapes to heap
	// @@ &bytes.Buffer literal escapes to heap
	c.Finalize()
	compressed := c.Bytes()
	fmt.Printf("%+v\n", compressed)
	fmt.Printf("%d bytes instead of %d bytes\n", len(compressed), len(data)*8)
	// @@ compressed escapes to heap
	d := compress.NewDecompressionBuffer(compressed, len(data))
	// @@ len(compressed) escapes to heap
	// @@ len(data) * 8 escapes to heap
	decompressed := d.Decompress()
	// @@ inlining call to compress.NewDecompressionBuffer
	fmt.Printf("Decompressed %f\n", decompressed)
}

// @@ decompressed escapes to heap
