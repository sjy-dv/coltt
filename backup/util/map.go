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

import "sort"

func CopyMap(m map[string]interface{}) map[string]interface{} {
	mapCopy := make(map[string]interface{})
	for k, v := range m {
		mapValue, ok := v.(map[string]interface{})
		if ok {
			mapCopy[k] = CopyMap(mapValue)
		} else {
			mapCopy[k] = v
		}
	}
	return mapCopy
}

func MapKeys(m map[string]interface{}, sorted bool, includeSubKeys bool) []string {
	keys := make([]string, 0, len(m))
	for key, value := range m {
		added := false
		if includeSubKeys {
			subMap, isMap := value.(map[string]interface{})
			if isMap {
				subFields := MapKeys(subMap, false, includeSubKeys)
				for _, subKey := range subFields {
					keys = append(keys, key+"."+subKey)
				}
				added = true
			}
		}

		if !added {
			keys = append(keys, key)
		}
	}

	if sorted {
		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})
	}

	return keys
}

func StringSliceToSet(s []string) map[string]bool {
	set := make(map[string]bool)
	for _, str := range s {
		set[str] = true
	}
	return set
}
