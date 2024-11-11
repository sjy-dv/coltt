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

package internal

import (
	"time"

	"github.com/google/orderedcode"
	"github.com/sjy-dv/nnv/backup/util"
)

func getEncodeValue(value interface{}) interface{} {
	if util.IsNumber(value) {
		return util.ToFloat64(value) // each number is encoded as float64
	}

	switch vType := value.(type) {
	case bool:
		return uint64(util.BoolToInt(vType))
	case time.Time:
		return uint64(vType.UnixNano())
	}
	return value
}

func orderedCodePrimitive(buf []byte, value interface{}, includeType bool) ([]byte, error) {
	var err error

	actualVal := getEncodeValue(value)
	if includeType {
		typeId := uint64(TypeId(value))
		buf, err = orderedcode.Append(buf, typeId)
		if err != nil {
			return nil, err
		}
	}

	if value == nil {
		return buf, nil
	}

	buf, err = orderedcode.Append(buf, actualVal)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func OrderedCode(buf []byte, v interface{}) ([]byte, error) {
	return orderedCode(buf, v, false)
}

func orderedCode(buf []byte, v interface{}, includeType bool) ([]byte, error) {
	switch vType := v.(type) {
	case map[string]interface{}:
		return orderedCodeObject(buf, vType)
	case []interface{}:
		return orderedCodeSlice(buf, vType)
	}
	return orderedCodePrimitive(buf, v, includeType)
}

func orderedCodeSlice(buf []byte, s []interface{}) ([]byte, error) {
	sliceEncoding := make([]byte, 0)
	for _, v := range s {
		var err error
		sliceEncoding, err = orderedCode(sliceEncoding, v, true)
		if err != nil {
			return nil, err
		}
	}
	return orderedcode.Append(buf, uint64(TypeId(s)), string(sliceEncoding))
}

func orderedCodeObject(buf []byte, o map[string]interface{}) ([]byte, error) {
	objEncoding := make([]byte, 0)
	for _, key := range util.MapKeys(o, true, false) {
		value := o[key]

		encoded, err := orderedcode.Append(objEncoding, key)
		if err != nil {
			return nil, err
		}

		objEncoding, err = orderedCode(encoded, value, true)
		if err != nil {
			return nil, err
		}
	}
	return orderedcode.Append(buf, uint64(TypeId(o)), string(objEncoding))
}
