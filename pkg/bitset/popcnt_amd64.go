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

//go:build !go1.9 && amd64 && !appengine
// +build !go1.9,amd64,!appengine

package bitset

// *** the following functions are defined in popcnt_amd64.s

//go:noescape

func hasAsm() bool

// useAsm is a flag used to select the GO or ASM implementation of the popcnt function
var useAsm = hasAsm()

//go:noescape

func popcntSliceAsm(s []uint64) uint64

//go:noescape

func popcntMaskSliceAsm(s, m []uint64) uint64

//go:noescape

func popcntAndSliceAsm(s, m []uint64) uint64

//go:noescape

func popcntOrSliceAsm(s, m []uint64) uint64

//go:noescape

func popcntXorSliceAsm(s, m []uint64) uint64

func popcntSlice(s []uint64) uint64 {
	if useAsm {
		return popcntSliceAsm(s)
	}
	return popcntSliceGo(s)
}

func popcntMaskSlice(s, m []uint64) uint64 {
	if useAsm {
		return popcntMaskSliceAsm(s, m)
	}
	return popcntMaskSliceGo(s, m)
}

func popcntAndSlice(s, m []uint64) uint64 {
	if useAsm {
		return popcntAndSliceAsm(s, m)
	}
	return popcntAndSliceGo(s, m)
}

func popcntOrSlice(s, m []uint64) uint64 {
	if useAsm {
		return popcntOrSliceAsm(s, m)
	}
	return popcntOrSliceGo(s, m)
}

func popcntXorSlice(s, m []uint64) uint64 {
	if useAsm {
		return popcntXorSliceAsm(s, m)
	}
	return popcntXorSliceGo(s, m)
}
