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

package conversion

import "encoding/binary"

// Converts a node id and a suffix to a byte slice key to be used in diskstore.
func NodeKey(id uint64, suffix byte) []byte {
	key := [10]byte{}
	key[0] = 'n'
	binary.LittleEndian.PutUint64(key[1:], id)
	key[9] = suffix
	return key[:]
}

// Checks if a given key and suffix is a valid a node key and returns the id.
func NodeIdFromKey(key []byte, suffix byte) (uint64, bool) {
	if len(key) != 10 || key[0] != 'n' || key[9] != suffix {
		return 0, false
	}
	return binary.LittleEndian.Uint64(key[1 : len(key)-1]), true
}
