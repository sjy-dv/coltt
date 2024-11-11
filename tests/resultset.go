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

// resultset.go
package edge

import (
	"fmt"
	"sync"
)

type Result struct {
	Similarity float32
	ID         ID
}

func (r Result) String() string {
	return fmt.Sprintf("(%d %0.4f)", r.ID, r.Similarity)
}

type ResultSet struct {
	inner sync.Mutex
	sims  []float32
	ids   []ID
	k     int
	valid int
}

func NewResultSet(topK int) *ResultSet {
	return &ResultSet{
		k:     topK,
		sims:  make([]float32, topK),
		ids:   make([]ID, topK),
		valid: 0,
	}
}

func (rs *ResultSet) Len() int {
	return len(rs.sims)
}

func (rs *ResultSet) ComputeRecall(baseline *ResultSet, at int) float64 {
	found := 0
	for _, v := range baseline.ids[:at] {
		for _, w := range rs.ids[:at] {
			if v == w {
				found += 1
			}
		}
	}
	return float64(found) / float64(at)
}

func (rs *ResultSet) String() string {
	return fmt.Sprint(rs.ToSlice())
}

func (rs *ResultSet) AddResult(id ID, sim float32) bool {
	// Do a quick check...
	if rs.valid == rs.k {
		// Bail if the last one beats us
		last := rs.sims[len(rs.sims)-1]
		if last > sim {
			return false
		}
	}
	rs.inner.Lock()
	defer rs.inner.Unlock()
	insert := 0
	found := false
	for insert != rs.k {
		// If we're building it out, then the new insertion point is at the end.
		if rs.valid <= insert {
			rs.valid += 1
			found = true
			break
		}
		if rs.ids[insert] == id {
			return true
		}
		if rs.sims[insert] < sim {
			found = true
			break
		}
		insert++
	}
	if !found {
		return false
	}
	copy(rs.sims[insert+1:], rs.sims[insert:])
	rs.sims[insert] = sim
	copy(rs.ids[insert+1:], rs.ids[insert:])
	rs.ids[insert] = id
	return true
}

func (rs *ResultSet) ToSlice() []*Result {
	out := make([]*Result, rs.valid)
	for i := range out {
		out[i] = &Result{
			Similarity: rs.sims[i],
			ID:         rs.ids[i],
		}
	}
	return out
}
