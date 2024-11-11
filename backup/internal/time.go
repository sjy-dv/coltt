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

	"github.com/sjy-dv/nnv/backup/util"
	"github.com/vmihailenco/msgpack/v5"
)

func init() {
	msgpack.RegisterExt(1, (*LocalizedTime)(nil))
}

type LocalizedTime struct {
	time.Time
}

var _ msgpack.Marshaler = (*LocalizedTime)(nil)
var _ msgpack.Unmarshaler = (*LocalizedTime)(nil)

func (tm *LocalizedTime) MarshalMsgpack() ([]byte, error) {
	return tm.GobEncode()
}

func (tm *LocalizedTime) UnmarshalMsgpack(b []byte) error {
	return tm.GobDecode(b)
}

func replaceTimes(v interface{}) interface{} {
	if t, isTime := v.(time.Time); isTime {
		return &LocalizedTime{t}
	}

	m, isMap := v.(map[string]interface{})
	if isMap {
		mapCopy := util.CopyMap(m)
		for k, v := range m {
			mapCopy[k] = replaceTimes(v)
		}
		return mapCopy
	}

	s, isSlice := v.([]interface{})
	if isSlice {
		sliceCopy := make([]interface{}, len(s))
		for i, v := range s {
			sliceCopy[i] = replaceTimes(v)
		}
		return sliceCopy
	}
	return v
}

func removeLocalizedTimes(v interface{}) interface{} {
	if t, isLTime := v.(*LocalizedTime); isLTime {
		return t.Time
	}

	m, isMap := v.(map[string]interface{})
	if isMap {
		for k, v := range m {
			m[k] = removeLocalizedTimes(v)
		}
	}

	s, isSlice := v.([]interface{})
	if isSlice {
		for i, v := range s {
			s[i] = replaceTimes(v)
		}
	}
	return v
}
