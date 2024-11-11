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

package index

import "github.com/sjy-dv/nnv/backup/internal"

type Range struct {
	Start, End                 interface{}
	StartIncluded, EndIncluded bool
}

func (r *Range) IsEmpty() bool {
	if (r.Start == nil && !r.StartIncluded && r.End != nil) || (r.End == nil && !r.EndIncluded && r.Start != nil) {
		return false
	}

	res := internal.Compare(r.Start, r.End)
	return (res > 0) || (res == 0 && !r.StartIncluded && !r.EndIncluded)
}

func (r *Range) IsNil() bool {
	return r.Start == nil && r.End == nil && r.StartIncluded && r.EndIncluded
}

func (r *Range) Intersect(r2 *Range) *Range {
	intersection := &Range{
		Start:         r.Start,
		End:           r.End,
		StartIncluded: r.StartIncluded,
		EndIncluded:   r.EndIncluded,
	}

	res := internal.Compare(r2.Start, intersection.Start)
	if res > 0 {
		intersection.Start = r2.Start
		intersection.StartIncluded = r2.StartIncluded
	} else if res == 0 {
		intersection.StartIncluded = intersection.StartIncluded && r2.StartIncluded
	} else if intersection.Start == nil {
		intersection.Start = r2.Start
		intersection.StartIncluded = r2.StartIncluded
	}

	res = internal.Compare(r2.End, intersection.End)
	if res < 0 {
		intersection.End = r2.End
		intersection.EndIncluded = r2.EndIncluded
	} else if res == 0 {
		intersection.EndIncluded = intersection.EndIncluded && r2.EndIncluded
	} else if intersection.End == nil {
		intersection.End = r2.End
		intersection.EndIncluded = r2.EndIncluded
	}
	return intersection
}
