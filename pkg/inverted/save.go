package inverted

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	roaring "github.com/RoaringBitmap/roaring/v2/roaring64"
)

func (idx *BitmapIndex) SerializeBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	idx.shardLock.RLock()
	shardCount := uint32(len(idx.Shards))
	if err := binary.Write(buf, binary.LittleEndian, shardCount); err != nil {
		idx.shardLock.RUnlock()
		return nil, fmt.Errorf("failed to write shard count: %v", err)
	}
	for key, shard := range idx.Shards {
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
			valueBytes := []byte(fmt.Sprintf("%v", value))
			valueLength := uint32(len(valueBytes))
			if err := binary.Write(buf, binary.LittleEndian, valueLength); err != nil {
				shard.rmu.RUnlock()
				idx.shardLock.RUnlock()
				return nil, fmt.Errorf("failed to write value length for %s:%v: %v", key, value, err)
			}
			if _, err := buf.Write(valueBytes); err != nil {
				shard.rmu.RUnlock()
				idx.shardLock.RUnlock()
				return nil, fmt.Errorf("failed to write value data for %s:%v: %v", key, value, err)
			}
			bitmapBytes, err := bitmap.ToBytes()
			if err != nil {
				shard.rmu.RUnlock()
				idx.shardLock.RUnlock()
				return nil, fmt.Errorf("failed to serialize bitmap for %s:%v: %v", key, value, err)
			}
			bitmapLength := uint32(len(bitmapBytes))
			if err := binary.Write(buf, binary.LittleEndian, bitmapLength); err != nil {
				shard.rmu.RUnlock()
				idx.shardLock.RUnlock()
				return nil, fmt.Errorf("failed to write bitmap length for %s:%v: %v", key, value, err)
			}
			if _, err := buf.Write(bitmapBytes); err != nil {
				shard.rmu.RUnlock()
				idx.shardLock.RUnlock()
				return nil, fmt.Errorf("failed to write bitmap data for %s:%v: %v", key, value, err)
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
			var valueLength uint32
			if err := binary.Read(buf, binary.LittleEndian, &valueLength); err != nil {
				return fmt.Errorf("failed to read value length for shard %s: %v", key, err)
			}
			valueBytes := make([]byte, valueLength)
			if _, err := io.ReadFull(buf, valueBytes); err != nil {
				return fmt.Errorf("failed to read value data for shard %s: %v", key, err)
			}
			value := string(valueBytes)
			var bitmapLength uint32
			if err := binary.Read(buf, binary.LittleEndian, &bitmapLength); err != nil {
				return fmt.Errorf("failed to read bitmap length for %s:%v: %v", key, value, err)
			}
			bitmapBytes := make([]byte, bitmapLength)
			if _, err := io.ReadFull(buf, bitmapBytes); err != nil {
				return fmt.Errorf("failed to read bitmap data for %s:%v: %v", key, value, err)
			}
			bitmap := roaring.New()
			if err := bitmap.UnmarshalBinary(bitmapBytes); err != nil {
				return fmt.Errorf("failed to unmarshal bitmap for %s:%v: %v", key, value, err)
			}
			shard.rmu.Lock()
			shard.ShardIndex[value] = bitmap
			shard.rmu.Unlock()
		}
	}
	return nil
}
