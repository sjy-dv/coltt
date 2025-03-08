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

package compresshelper

import (
	"math"
	"strconv"
)

type Float8 uint8

type F8Precision int

const (

	// PrecisionExact is for non-subnormals that don't drop bits during conversion.
	// All of these can round-trip.  Should always convert to float8.
	F8PrecisionExact F8Precision = iota

	// PrecisionUnknown is for subnormals that don't drop bits during conversion but
	// not all of these can round-trip so precision is unknown without more effort.
	// Only 2046 of these can round-trip and the rest cannot round-trip.
	F8PrecisionUnknown

	// PrecisionInexact is for dropped significand bits and cannot round-trip.
	// Some of these are subnormals. Cannot round-trip float32->float8->float32.
	F8PrecisionInexact

	// PrecisionUnderflow is for Underflows. Cannot round-trip float32->float8->float32.
	F8PrecisionUnderflow

	// PrecisionOverflow is for Overflows. Cannot round-trip float32->float8->float32.
	F8PrecisionOverflow
)

// Float8 represents IEEE 754 half-precision floating-point numbers (binary8).

// Precision indicates whether the conversion to Float8 is
// exact, subnormal without dropped bits, inexact, underflow, or overflow.

// PrecisionFromfloat32 returns Precision without performing
// the conversion.  Conversions from both Infinity and NaN
// values will always report PrecisionExact even if NaN payload
// or NaN-Quiet-Bit is lost. This function is kept simple to
// allow inlining and run < 0.5 ns/op, to serve as a fast filter.

func F8PrecisionFromfloat32(f32 float32) F8Precision {
	u32 := math.Float32bits(f32)

	if u32 == 0 || u32 == 0x800000 {
		// +- zero will always be exact conversion
		return F8PrecisionExact
	}

	const COEFMASK uint32 = 0x7fffff // 23 least significant bits
	const EXPSHIFT uint32 = 23
	const EXPBIAS uint32 = 127
	const EXPMASK uint32 = uint32(0xff) << EXPSHIFT
	const DROPMASK uint32 = COEFMASK >> 10

	exp := int32(((u32 & EXPMASK) >> EXPSHIFT) - EXPBIAS)
	coef := u32 & COEFMASK

	if exp == 128 {
		// +- infinity or NaN
		// apps may want to do extra checks for NaN separately
		return F8PrecisionExact
	}

	// https://en.wikipedia.org/wiki/Half-precision_floating-point_format says,
	// "Decimals between 2^−24 (minimum positive subnormal) and 2^−14 (maximum subnormal): fixed interval 2^−24"
	if exp < -24 {
		return F8PrecisionUnderflow
	}
	if exp > 15 {
		return F8PrecisionOverflow
	}
	if (coef & DROPMASK) != uint32(0) {
		// these include subnormals and non-subnormals that dropped bits
		return F8PrecisionInexact
	}

	if exp < -14 {
		// Subnormals. Caller may want to test these further.
		// There are 2046 subnormals that can successfully round-trip f32->f8->f32
		// and 20 of those 2046 have 32-bit input coef == 0.
		// RFC 7049 and 7049bis Draft 12 don't precisely define "preserves value"
		// so some protocols and libraries will choose to handle subnormals differently
		// when deciding to encode them to CBOR float32 vs float8.
		return F8PrecisionUnknown
	}

	return F8PrecisionExact
}

// Frombits returns the float8 number corresponding to the IEEE 754 binary8
// representation u8, with the sign bit of u8 and the result in the same bit
// position. Frombits(Bits(x)) == x.
func F8Frombits(u8 uint8) Float8 {
	return Float8(u8)
}

// Fromfloat32 returns a Float8 value converted from f32. Conversion uses
// IEEE default rounding (nearest int, with ties to even).
func F8Fromfloat32(f32 float32) Float8 {
	return Float8(f32bitsToF8bits(math.Float32bits(f32)))
}

// F8ErrInvalidNaNValue indicates a NaN was not received.
const F8ErrInvalidNaNValue = Float8Error("float8: invalid NaN value, expected IEEE 754 NaN")

type Float8Error string

func (e Float8Error) Error() string { return string(e) }

// FromNaN32ps converts nan to IEEE binary8 NaN while preserving both
// signaling and payload. Unlike Fromfloat32(), which can only return
// qNaN because it sets quiet bit = 1, this can return both sNaN and qNaN.
// If the result is infinity (sNaN with empty payload), then the
// lowest bit of payload is set to make the result a NaN.
// Returns F8F8ErrInvalidNaNValue and 0x7c (sNaN) if nan isn't IEEE 754 NaN.
// This function was kept simple to be able to inline.
func F8FromNaN32ps(nan float32) (Float8, error) {
	const SNAN = Float8(uint8(0x7c)) // signaling NaN

	u32 := math.Float32bits(nan)
	sign := u32 & 0x800000
	exp := u32 & 0x7f800000
	coef := u32 & 0x007fffff

	if (exp != 0x7f800000) || (coef == 0) {
		return SNAN, F8ErrInvalidNaNValue
	}

	u8 := uint8((sign >> 8) | uint32(0x7c) | (coef >> 13))

	if (u8 & 0x03) == 0 {
		// result became infinity, make it NaN by setting lowest bit in payload
		u8 |= 0x0001
	}

	return Float8(u8), nil
}

// NaN returns a Float8 of IEEE 754 binary8 not-a-number (NaN).
// Returned NaN value 0x7e01 has all exponent bits = 1 with the
// first and last bits = 1 in the significand. This is consistent
// with Go's 64-bit math.NaN(). Canonical CBOR in RFC 7049 uses 0x7e00.
func F8NaN() Float8 {
	return Float8(0x7e)
}

// Inf returns a Float8 with an infinity value with the specified sign.
// A sign >= returns positive infinity.
// A sign < 0 returns negative infinity.
func F8Inf(sign int) Float8 {
	if sign >= 0 {
		return Float8(0x7c)
	}
	return Float8(0x80 | 0x7c)
}

// Float32 returns a float32 converted from f (Float8).
// This is a lossless conversion.
func (f Float8) Float32() float32 {
	u32 := F8bitsToF32bits(uint8(f))
	return math.Float32frombits(u32)
}

// Bits returns the IEEE 754 binary8 representation of f, with the sign bit
// of f and the result in the same bit position. Bits(Frombits(x)) == x.
func (f Float8) Bits() uint8 {
	return uint8(f)
}

// IsNaN reports whether f is an IEEE 754 binary8 “not-a-number” value.
func (f Float8) IsNaN() bool {
	return (f&0x7c == 0x7c) && (f&0x03 != 0)
}

// IsQuietNaN reports whether f is a quiet (non-signaling) IEEE 754 binary8
// “not-a-number” value.
func (f Float8) IsQuietNaN() bool {
	return (f&0x7c == 0x7c) && (f&0x03 != 0) && (f&0x02 != 0)
}

// IsInf reports whether f is an infinity (inf).
// A sign > 0 reports whether f is positive inf.
// A sign < 0 reports whether f is negative inf.
// A sign == 0 reports whether f is either inf.
func (f Float8) IsInf(sign int) bool {
	return ((f == 0x7c) && sign >= 0) ||
		(f == 0xfc && sign <= 0)
}

// IsFinite returns true if f is neither infinite nor NaN.
func (f Float8) IsFinite() bool {
	return (uint8(f) & uint8(0x7c)) != uint8(0x7c)
}

// IsNormal returns true if f is neither zero, infinite, subnormal, or NaN.
func (f Float8) IsNormal() bool {
	exp := uint8(f) & uint8(0x7c)
	return (exp != uint8(0x7c)) && (exp != 0)
}

// Signbit reports whether f is negative or negative zero.
func (f Float8) Signbit() bool {
	return (uint8(f) & uint8(0x80)) != 0
}

// String satisfies the fmt.Stringer interface.
func (f Float8) String() string {
	return strconv.FormatFloat(float64(f.Float32()), 'f', -1, 32)
}

// bf8bitsToF32bits returns uint32 (float32 bits) converted from specified uint8.
func F8bitsToF32bits(in uint8) uint32 {
	// All 65536 conversions with this were confirmed to be correct
	// by Montgomery Edwards⁴⁴⁸ (github.com/x448).

	sign := uint32(in&0x80) << 8  // sign for 32-bit
	exp := uint32(in&0x7c) >> 10  // exponenent for 8-bit
	coef := uint32(in&0x03) << 13 // significand for 32-bit

	if exp == 0x1f {
		if coef == 0 {
			// infinity
			return sign | 0x7f800000 | coef
		}
		// NaN
		return sign | 0x7fc00000 | coef
	}

	if exp == 0 {
		if coef == 0 {
			// zero
			return sign
		}

		// normalize subnormal numbers
		exp++
		for coef&0x7f800000 == 0 {
			coef <<= 1
			exp--
		}
		coef &= 0x007fffff
	}

	return sign | ((exp + (0x7f - 0xf)) << 23) | coef
}

// f32bitsToF8bits returns uint8 (Float8 bits) converted from the specified float32.
// Conversion rounds to nearest integer with ties to even.
func f32bitsToF8bits(u32 uint32) uint8 {
	sign := u32 & 0x800000
	exp := u32 & 0x7f800000
	coef := u32 & 0x007fffff

	if exp == 0x7f800000 {
		// NaN or Infinity
		nanBit := uint32(0)
		if coef != 0 {
			nanBit = uint32(0x0200)
		}
		return uint8((sign >> 8) | uint32(0x7c) | nanBit | (coef >> 13))
	}

	halfSign := sign >> 8

	unbiasedExp := int32(exp>>23) - 127
	halfExp := unbiasedExp + 15

	if halfExp >= 0x1f {
		return uint8(halfSign | uint32(0x7c))
	}

	if halfExp <= 0 {
		if 14-halfExp > 24 {
			return uint8(halfSign)
		}
		c := coef | uint32(0x00800000)
		halfCoef := c >> uint32(14-halfExp)
		roundBit := uint32(1) << uint32(13-halfExp)
		if (c&roundBit) != 0 && (c&(3*roundBit-1)) != 0 {
			halfCoef++
		}
		return uint8(halfSign | halfCoef)
	}

	uHalfExp := uint32(halfExp) << 10
	halfCoef := coef >> 13
	roundBit := uint32(0x00001000)
	if (coef&roundBit) != 0 && (coef&(3*roundBit-1)) != 0 {
		return uint8((halfSign | uHalfExp | halfCoef) + 1)
	}
	return uint8(halfSign | uHalfExp | halfCoef)
}
