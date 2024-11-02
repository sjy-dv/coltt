package index

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/RoaringBitmap/roaring"
)

func (idx *BitmapIndex) ValidateIndex() error {
	idx.shardLock.RLock()
	defer idx.shardLock.RUnlock()

	for key, shard := range idx.Shards {
		shard.rmu.RLock()
		for val, bitamp := range shard.ShardIndex {
			if bitamp == nil {
				shard.rmu.RUnlock()
				return fmt.Errorf("bitmap is nil for %s:%s", key, val)
			}
		}
		shard.rmu.RUnlock()
	}
	return nil
}

func (idx *BitmapIndex) SerializeBinary(filename string) error {

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filename, err)
	}
	defer file.Close()

	idx.shardLock.RLock()
	idxCount := uint32(len(idx.Shards))

	if err := binary.Write(file, binary.LittleEndian, idxCount); err != nil {
		idx.shardLock.RUnlock()
		return fmt.Errorf("failed to write tag key count: %v", err)
	}

	for key, shard := range idx.Shards {
		keyBytes := []byte(key)
		keyLength := uint32(len(keyBytes))
		if err := binary.Write(file, binary.LittleEndian, keyLength); err != nil {
			idx.shardLock.RUnlock()
			return fmt.Errorf("failed to write tag key length for %s: %v", key, err)
		}
		if _, err := file.Write(keyBytes); err != nil {
			idx.shardLock.RUnlock()
			return fmt.Errorf("failed to write tag key data for %s: %v", key, err)
		}

		shard.rmu.RLock()
		valueCount := uint32(len(shard.ShardIndex))
		if err := binary.Write(file, binary.LittleEndian, valueCount); err != nil {
			shard.rmu.RUnlock()
			idx.shardLock.RUnlock()
			return fmt.Errorf("failed to write value count for key %s: %v", key, err)
		}

		for value, bitmap := range shard.ShardIndex {
			valueBytes := []byte(value)
			valueLength := uint32(len(valueBytes))
			if err := binary.Write(file, binary.LittleEndian, valueLength); err != nil {
				shard.rmu.RUnlock()
				idx.shardLock.RUnlock()
				return fmt.Errorf("failed to write tag value length for %s:%s: %v", key, value, err)
			}
			if _, err := file.Write(valueBytes); err != nil {
				shard.rmu.RUnlock()
				idx.shardLock.RUnlock()
				return fmt.Errorf("failed to write tag value data for %s:%s: %v", key, value, err)
			}

			bitmapBytes, err := bitmap.ToBytes()
			if err != nil {
				shard.rmu.RUnlock()
				idx.shardLock.RUnlock()
				return fmt.Errorf("failed to serialize bitmap for %s:%s: %v", key, value, err)
			}
			bitmapLength := uint32(len(bitmapBytes))
			if err := binary.Write(file, binary.LittleEndian, bitmapLength); err != nil {
				shard.rmu.RUnlock()
				idx.shardLock.RUnlock()
				return fmt.Errorf("failed to write bitmap length for %s:%s: %v", key, value, err)
			}
			if _, err := file.Write(bitmapBytes); err != nil {
				shard.rmu.RUnlock()
				idx.shardLock.RUnlock()
				return fmt.Errorf("failed to write bitmap data for %s:%s: %v", key, value, err)
			}
		}
		shard.rmu.RUnlock()
	}

	idx.shardLock.RUnlock()
	return nil
}

func (idx *BitmapIndex) DeserializeBinary(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		if err == os.ErrNotExist {
			return nil
		}
		return fmt.Errorf("failed to open file %s: %v", filename, err)
	}
	defer file.Close()

	var tagKeyCount uint32
	if err := binary.Read(file, binary.LittleEndian, &tagKeyCount); err != nil {
		return fmt.Errorf("failed to read tag key count: %v", err)
	}

	for i := uint32(0); i < tagKeyCount; i++ {
		var keyLength uint32
		if err := binary.Read(file, binary.LittleEndian, &keyLength); err != nil {
			return fmt.Errorf("failed to read tag key length: %v", err)
		}
		keyBytes := make([]byte, keyLength)
		if _, err := io.ReadFull(file, keyBytes); err != nil {
			return fmt.Errorf("failed to read tag key data: %v", err)
		}
		key := string(keyBytes)

		var valueCount uint32
		if err := binary.Read(file, binary.LittleEndian, &valueCount); err != nil {
			return fmt.Errorf("failed to read value count for key %s: %v", key, err)
		}

		shard := idx.getShard(key)

		for j := uint32(0); j < valueCount; j++ {
			var valueLength uint32
			if err := binary.Read(file, binary.LittleEndian, &valueLength); err != nil {
				return fmt.Errorf("failed to read tag value length for key %s: %v", key, err)
			}
			valueBytes := make([]byte, valueLength)
			if _, err := io.ReadFull(file, valueBytes); err != nil {
				return fmt.Errorf("failed to read tag value data for key %s: %v", key, err)
			}
			value := string(valueBytes)

			var bitmapLength uint32
			if err := binary.Read(file, binary.LittleEndian, &bitmapLength); err != nil {
				return fmt.Errorf("failed to read bitmap length for key %s, value %s: %v", key, value, err)
			}
			bitmapBytes := make([]byte, bitmapLength)
			if _, err := io.ReadFull(file, bitmapBytes); err != nil {
				return fmt.Errorf("failed to read bitmap data for key %s, value %s: %v", key, value, err)
			}

			bitmap := roaring.New()
			if err := bitmap.UnmarshalBinary(bitmapBytes); err != nil {
				return fmt.Errorf("failed to unmarshal bitmap for key %s, value %s: %v", key, value, err)
			}

			shard.rmu.Lock()
			shard.ShardIndex[value] = bitmap
			shard.rmu.Unlock()
		}
	}

	return nil
}
