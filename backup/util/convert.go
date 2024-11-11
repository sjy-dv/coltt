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

package util

func IsNumber(v interface{}) bool {
	switch v.(type) {
	case int, uint, uint8, uint16, uint32, uint64,
		int8, int16, int32, int64, float32, float64:
		return true
	default:
		return false
	}
}

func ToFloat64(v interface{}) float64 {
	switch vType := v.(type) {
	case uint32:
		return float64(vType)
	case uint64:
		return float64(vType)
	case int64:
		return float64(vType)
	case float64:
		return vType
	}
	panic("not a number")
}

func ToInt64(v interface{}) int64 {
	switch vType := v.(type) {
	case uint64:
		return int64(vType)
	case int64:
		return vType
	}
	panic("not a number")
}

func BoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
