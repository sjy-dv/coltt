// Licensed to sjy-dv under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. sjy-dv licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package bitset

func select64(w uint64, j uint) uint {
	seen := 0
	// Divide 64bit
	part := w & 0xFFFFFFFF
	n := uint(popcount(part))
	if n <= j {
		part = w >> 32
		seen += 32
		j -= n
	}
	ww := part

	// Divide 32bit
	part = ww & 0xFFFF

	n = uint(popcount(part))
	if n <= j {
		part = ww >> 16
		seen += 16
		j -= n
	}
	ww = part

	// Divide 16bit
	part = ww & 0xFF
	n = uint(popcount(part))
	if n <= j {
		part = ww >> 8
		seen += 8
		j -= n
	}
	ww = part

	// Lookup in final byte
	counter := 0
	for ; counter < 8; counter++ {
		j -= uint((ww >> counter) & 1)
		if j+1 == 0 {
			break
		}
	}
	return uint(seen + counter)
}
