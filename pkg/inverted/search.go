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

package inverted

import (
	"fmt"

	roaring "github.com/RoaringBitmap/roaring/v2/roaring64"
)

func (idx *BitmapIndex) evaluateSingleFilter(f *Filter) (*roaring.Bitmap, error) {
	shard := idx.getShard(f.IndexName)
	result := roaring.New()
	shard.rmu.RLock()
	defer shard.rmu.RUnlock()

	if f.Op == OpEqual {
		if bm, exists := shard.ShardIndex[f.Value]; exists {
			result.Or(bm)
		}
	} else {
		for key, bm := range shard.ShardIndex {
			match, err := satisfiesOp(f, key)
			if err != nil {
				return nil, err
			}
			if match {
				result.Or(bm)
			}
		}
	}
	return result, nil
}

func (idx *BitmapIndex) evaluateCompositeFilter(cf *CompositeFilter) (*roaring.Bitmap, error) {
	var result *roaring.Bitmap
	if cf.Op == LogicalAnd {
		for _, expr := range cf.Expressions {
			bm, err := idx.evaluateFilterExpression(expr)
			if err != nil {
				return nil, err
			}
			if result == nil {
				result = bm.Clone()
			} else {
				result.And(bm)
			}
		}
	} else if cf.Op == LogicalOr {
		result = roaring.New()
		for _, expr := range cf.Expressions {
			bm, err := idx.evaluateFilterExpression(expr)
			if err != nil {
				return nil, err
			}
			result.Or(bm)
		}
	} else {
		return nil, fmt.Errorf("unsupported composite op")
	}
	return result, nil
}

func (idx *BitmapIndex) evaluateFilterExpression(expr *FilterExpression) (*roaring.Bitmap, error) {
	if expr.Single != nil {
		return idx.evaluateSingleFilter(expr.Single)
	} else if expr.Composite != nil {
		return idx.evaluateCompositeFilter(expr.Composite)
	}
	return nil, fmt.Errorf("empty filter expression")
}

func (idx *BitmapIndex) SearchSingleFilter(f *Filter) ([]uint64, error) {
	bm, err := idx.evaluateSingleFilter(f)
	if err != nil {
		return nil, err
	}
	return bm.ToArray(), nil
}

func (idx *BitmapIndex) SearchMultiFilter(filters []*Filter) ([]uint64, error) {
	var result *roaring.Bitmap
	first := true
	for _, filter := range filters {
		bm, err := idx.evaluateSingleFilter(filter)
		if err != nil {
			return nil, err
		}
		if first {
			result = bm.Clone()
			first = false
		} else {
			result.And(bm)
		}
	}
	return result.ToArray(), nil
}
func (idx *BitmapIndex) SearchWithExpression(expr *FilterExpression) ([]uint64, error) {
	bm, err := idx.evaluateFilterExpression(expr)
	if err != nil {
		return nil, err
	}
	return bm.ToArray(), nil
}
