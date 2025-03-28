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

//go:build go1.9
// +build go1.9

package bitset

import "math/bits"

func popcntSlice(s []uint64) uint64 {
	var cnt int
	for _, x := range s {
		cnt += bits.OnesCount64(x)
	}
	return uint64(cnt)
}

func popcntMaskSlice(s, m []uint64) uint64 {
	var cnt int
	// this explicit check eliminates a bounds check in the loop
	if len(m) < len(s) {
		panic("mask slice is too short")
	}
	for i := range s {
		cnt += bits.OnesCount64(s[i] &^ m[i])
	}
	return uint64(cnt)
}

func popcntAndSlice(s, m []uint64) uint64 {
	var cnt int
	// this explicit check eliminates a bounds check in the loop
	if len(m) < len(s) {
		panic("mask slice is too short")
	}
	for i := range s {
		cnt += bits.OnesCount64(s[i] & m[i])
	}
	return uint64(cnt)
}

func popcntOrSlice(s, m []uint64) uint64 {
	var cnt int
	// this explicit check eliminates a bounds check in the loop
	if len(m) < len(s) {
		panic("mask slice is too short")
	}
	for i := range s {
		cnt += bits.OnesCount64(s[i] | m[i])
	}
	return uint64(cnt)
}

func popcntXorSlice(s, m []uint64) uint64 {
	var cnt int
	// this explicit check eliminates a bounds check in the loop
	if len(m) < len(s) {
		panic("mask slice is too short")
	}
	for i := range s {
		cnt += bits.OnesCount64(s[i] ^ m[i])
	}
	return uint64(cnt)
}
