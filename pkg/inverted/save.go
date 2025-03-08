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
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	roaring "github.com/RoaringBitmap/roaring/v2/roaring64"
)

func writeValue(buf *bytes.Buffer, val interface{}) error {
	switch v := val.(type) {
	case int, int8, int16, int32, int64:
		// Type tag 0: int64
		if err := buf.WriteByte(0); err != nil {
			return err
		}
		var num int64
		switch t := v.(type) {
		case int:
			num = int64(t)
		case int8:
			num = int64(t)
		case int16:
			num = int64(t)
		case int32:
			num = int64(t)
		case int64:
			num = t
		}
		return binary.Write(buf, binary.BigEndian, num)
	case float32, float64:
		// Type tag 1: float64
		if err := buf.WriteByte(1); err != nil {
			return err
		}
		var num float64
		switch t := v.(type) {
		case float32:
			num = float64(t)
		case float64:
			num = t
		}
		return binary.Write(buf, binary.BigEndian, num)
	case string:
		// Type tag 2: string
		if err := buf.WriteByte(2); err != nil {
			return err
		}
		strBytes := []byte(v)
		if len(strBytes) > 65535 {
			return fmt.Errorf("string too long: %s", v)
		}
		if err := binary.Write(buf, binary.BigEndian, uint16(len(strBytes))); err != nil {
			return err
		}
		_, err := buf.Write(strBytes)
		return err
	case bool:
		// Type tag 3: bool
		if err := buf.WriteByte(3); err != nil {
			return err
		}
		var b byte = 0
		if v {
			b = 1
		}
		return buf.WriteByte(b)
	default:
		return fmt.Errorf("unsupported metadata type: %T", v)
	}
}

func readValue(buf *bytes.Reader) (interface{}, error) {
	tag, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}
	switch tag {
	case 0: // int64
		var num int64
		if err := binary.Read(buf, binary.BigEndian, &num); err != nil {
			return nil, err
		}
		return num, nil
	case 1: // float64
		var num float64
		if err := binary.Read(buf, binary.BigEndian, &num); err != nil {
			return nil, err
		}
		return num, nil
	case 2: // string
		var strLen uint16
		if err := binary.Read(buf, binary.BigEndian, &strLen); err != nil {
			return nil, err
		}
		strBytes := make([]byte, strLen)
		if _, err := io.ReadFull(buf, strBytes); err != nil {
			return nil, err
		}
		return string(strBytes), nil
	case 3: // bool
		b, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		return b != 0, nil
	default:
		return nil, fmt.Errorf("unsupported metadata type tag: %d", tag)
	}
}

func (idx *BitmapIndex) SerializeBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	idx.shardLock.RLock()
	shardCount := uint32(len(idx.Shards))
	if err := binary.Write(buf, binary.LittleEndian, shardCount); err != nil {
		idx.shardLock.RUnlock()
		return nil, fmt.Errorf("failed to write shard count: %v", err)
	}
	for key, shard := range idx.Shards {
		// Shard key (string) 저장
		keyBytes := []byte(key)
		keyLength := uint32(len(keyBytes))
		if err := binary.Write(buf, binary.LittleEndian, keyLength); err != nil {
			idx.shardLock.RUnlock()
			return nil, fmt.Errorf("failed to write shard key length for %s: %v", key, err)
		}
		if _, err := buf.Write(keyBytes); err != nil {
			idx.shardLock.RUnlock()
			return nil, fmt.Errorf("failed to write shard key data for %s: %v", key, err)
		}
		shard.rmu.RLock()
		valueCount := uint32(len(shard.ShardIndex))
		if err := binary.Write(buf, binary.LittleEndian, valueCount); err != nil {
			shard.rmu.RUnlock()
			idx.shardLock.RUnlock()
			return nil, fmt.Errorf("failed to write value count for shard %s: %v", key, err)
		}
		for value, bitmap := range shard.ShardIndex {
			// 값(value)을 custom writeValue 함수를 이용하여 저장
			if err := writeValue(buf, value); err != nil {
				shard.rmu.RUnlock()
				idx.shardLock.RUnlock()
				return nil, fmt.Errorf("failed to write value for shard %s: %v", key, err)
			}
			bitmapBytes, err := bitmap.ToBytes()
			if err != nil {
				shard.rmu.RUnlock()
				idx.shardLock.RUnlock()
				return nil, fmt.Errorf("failed to serialize bitmap for shard %s: %v", key, err)
			}
			bitmapLength := uint32(len(bitmapBytes))
			if err := binary.Write(buf, binary.LittleEndian, bitmapLength); err != nil {
				shard.rmu.RUnlock()
				idx.shardLock.RUnlock()
				return nil, fmt.Errorf("failed to write bitmap length for shard %s: %v", key, err)
			}
			if _, err := buf.Write(bitmapBytes); err != nil {
				shard.rmu.RUnlock()
				idx.shardLock.RUnlock()
				return nil, fmt.Errorf("failed to write bitmap data for shard %s: %v", key, err)
			}
		}
		shard.rmu.RUnlock()
	}
	idx.shardLock.RUnlock()
	return buf.Bytes(), nil
}

func (idx *BitmapIndex) DeserializeBinary(data []byte) error {
	buf := bytes.NewReader(data)
	var shardCount uint32
	if err := binary.Read(buf, binary.LittleEndian, &shardCount); err != nil {
		return fmt.Errorf("failed to read shard count: %v", err)
	}
	for i := uint32(0); i < shardCount; i++ {
		var keyLength uint32
		if err := binary.Read(buf, binary.LittleEndian, &keyLength); err != nil {
			return fmt.Errorf("failed to read shard key length: %v", err)
		}
		keyBytes := make([]byte, keyLength)
		if _, err := io.ReadFull(buf, keyBytes); err != nil {
			return fmt.Errorf("failed to read shard key data: %v", err)
		}
		key := string(keyBytes)
		var valueCount uint32
		if err := binary.Read(buf, binary.LittleEndian, &valueCount); err != nil {
			return fmt.Errorf("failed to read value count for shard %s: %v", key, err)
		}
		shard := idx.getShard(key)
		for j := uint32(0); j < valueCount; j++ {
			// readValue 함수를 사용해서 값(value)을 복원
			val, err := readValue(buf)
			if err != nil {
				return fmt.Errorf("failed to read value for shard %s: %v", key, err)
			}
			var bitmapLength uint32
			if err := binary.Read(buf, binary.LittleEndian, &bitmapLength); err != nil {
				return fmt.Errorf("failed to read bitmap length for %s:%v: %v", key, val, err)
			}
			bitmapBytes := make([]byte, bitmapLength)
			if _, err := io.ReadFull(buf, bitmapBytes); err != nil {
				return fmt.Errorf("failed to read bitmap data for %s:%v: %v", key, val, err)
			}
			bitmap := roaring.New()
			if err := bitmap.UnmarshalBinary(bitmapBytes); err != nil {
				return fmt.Errorf("failed to unmarshal bitmap for %s:%v: %v", key, val, err)
			}
			shard.rmu.Lock()
			shard.ShardIndex[val] = bitmap
			shard.rmu.Unlock()
		}
	}
	return nil
}
