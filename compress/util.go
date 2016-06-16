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

// MaxUint32 is the largest unsigned integer.
const MaxUint32 = ^uint32(0)

func leadingZeros64(x uint64) uint32 {
	var upper uint32
	var lower uint32
	lower = (uint32)(x)
	upper = (uint32)(x >> 32)

	zeros := leadingZeros(upper)
	if zeros == 32 {
		// @@ inlining call to leadingZeros
		zeros += leadingZeros(lower)
	}
	// @@ inlining call to leadingZeros
	return zeros
}

func leadingZeros(x uint32) uint32 {
	n := uint32(0)
	// @@ can inline leadingZeros
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

func nthLowestBit(n uint32, value uint64) bool {
	return 1 == (value>>n)&1
	// @@ can inline nthLowestBit
}
