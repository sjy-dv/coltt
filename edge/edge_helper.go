package edge

import (
	"fmt"
	"math"
	"os"

	"github.com/sjy-dv/coltt/gen/protoc/v3/edgeproto"
	"github.com/vmihailenco/msgpack/v5"
)

func errorWrap(errMsg string) *edgeproto.Error {
	return &edgeproto.Error{
		ErrorMessage: errMsg,
		ErrorCode:    edgeproto.ErrorCode_INTERNAL_FUNC_ERROR,
	}
}

func stateTrueHelper(collectionName string) {
	stateManager.checker.cecLock.Lock()
	defer stateManager.checker.cecLock.Unlock()
	stateManager.auth.authLock.Lock()
	defer stateManager.auth.authLock.Unlock()
	stateManager.checker.collections[collectionName] = true
	stateManager.auth.collections[collectionName] = true
}

func stateRegistHelper(collectionName string) {
	stateManager.checker.cecLock.Lock()
	defer stateManager.checker.cecLock.Unlock()
	stateManager.checker.collections[collectionName] = true
}

func stateFalseHelper(collectionName string) {
	stateManager.auth.authLock.Lock()
	defer stateManager.auth.authLock.Unlock()
	stateManager.auth.collections[collectionName] = false
}

func stateDestroyHelper(collectionName string) {
	stateManager.auth.authLock.Lock()
	defer stateManager.auth.authLock.Unlock()
	stateManager.checker.cecLock.Lock()
	defer stateManager.checker.cecLock.Unlock()
	delete(stateManager.auth.collections, collectionName)
	delete(stateManager.checker.collections, collectionName)
}

func collectionStatusHelper(collectionName string) error {
	if !hasCollection(collectionName) {
		return fmt.Errorf(ErrCollectionNotFound, collectionName)
	}
	if !alreadyLoadCollection(collectionName) {
		return fmt.Errorf(ErrCollectionNotLoad, collectionName)
	}
	return nil
}

func protoDistQuantizationHelper(dist edgeproto.Distance, quantiz edgeproto.Quantization) (string, string) {
	return func() string {
			if dist == edgeproto.Distance_Cosine {
				return COSINE
			}
			return EUCLIDEAN
		}(), func() string {
			if quantiz == edgeproto.Quantization_F16 {
				return F16_QUANTIZATION
			}
			if quantiz == edgeproto.Quantization_F8 {
				return F8_QUANTIZATION
			}
			if quantiz == edgeproto.Quantization_BF16 {
				return BF16_QUANTIZATION
			}
			return NONE_QAUNTIZATION
		}()

}

func allremover(collectionName string) {
	os.Remove(fmt.Sprintf(edgeData, collectionName))
	os.Remove(fmt.Sprintf(edgeIndex, collectionName))
	os.Remove(fmt.Sprintf(edgeVector, collectionName))
	os.Remove(fmt.Sprintf(edgeConfig, collectionName))
}

func (helper *Edge) memFree(collectionName string) {
	helper.Datas.Del(collectionName)

	indexdb.indexLock.Lock()
	delete(indexdb.indexes, collectionName)
	indexdb.indexLock.Unlock()

	helper.VectorStore.slock.Lock()
	delete(helper.VectorStore.Space, collectionName)
	helper.VectorStore.slock.Unlock()
}

func scoreHelper(score float32, dist string) float32 {
	if dist == COSINE {
		return ((2 - score) / 2) * 100
	}
	return float32(math.Max(0, float64(100-score)))
}

func (helper *Edge) diskClear(collectionName string) {
	var err error
	err = helper.VectorStore.DropCollection(collectionName)
	if err != nil {
		//
		return
	}
	err = indexdb.DropIndex(collectionName)
	if err != nil {
		//
		return
	}
	sep := fmt.Sprintf("%s_", collectionName)
	// add commit -trace log
	helper.Disk.AscendKeys([]byte(sep), true, func(k []byte) (bool, error) {
		err := helper.Disk.Delete(k)
		if err != nil {
			return false, err
		}
		return true, nil
	})
	allremover(collectionName)
}

func (helper *Edge) saveCollection(collectionName string) error {
	ok, err := helper.Disk.Exist([]byte(diskColList))
	if err != nil {
		return err
	}
	if ok {
		gb, err := helper.Disk.Get([]byte(diskColList))
		if err != nil {
			return err
		}
		data := make([]string, 0)
		err = msgpack.Unmarshal(gb, &data)
		if err != nil {
			return err
		}
		data = append(data, collectionName)
		byts, err := msgpack.Marshal(data)
		if err != nil {
			return err
		}
		err = helper.Disk.Put([]byte(diskColList), byts)
		if err != nil {
			return err
		}
	} else {
		data := []string{collectionName}
		bytes, err := msgpack.Marshal(data)
		if err != nil {
			return err
		}
		err = helper.Disk.Put([]byte(diskColList), bytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func (helper *Edge) ChkValidDimensionality(collectionName string, dim int32) error {
	collection, _ := helper.Datas.Get(collectionName)
	if collection.dim != dim {
		return fmt.Errorf("Err Collection %s expects %d dimensions, but has %d dimensions", collectionName, collection.dim, dim)
	}
	return nil
}

func (helper *Edge) removeCollection(collectionName string) error {
	gb, err := helper.Disk.Get([]byte(diskColList))
	if err != nil {
		return err
	}
	data := make([]string, 0)
	err = msgpack.Unmarshal(gb, &data)
	if err != nil {
		return err
	}

	newData := make([]string, 0)
	for _, d := range data {
		if d != collectionName {
			newData = append(newData, d)
		}
	}
	byts, err := msgpack.Marshal(newData)
	if err != nil {
		return err
	}
	err = helper.Disk.Put([]byte(diskColList), byts)
	if err != nil {
		return err
	}
	return nil
}

func (helper *Edge) RegistCollectionStManager() error {
	ok, err := helper.Disk.Exist([]byte(diskColList))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	gb, err := helper.Disk.Get([]byte(diskColList))
	if err != nil {
		return err
	}
	data := make([]string, 0)
	err = msgpack.Unmarshal(gb, &data)
	if err != nil {
		return err
	}
	for _, col := range data {
		stateRegistHelper(col)
	}
	return nil
}

func (helper *Edge) failIsDelete(id uint64, collectionName string, metadata map[string]any) {
	indexdb.indexes[collectionName].Remove(id, metadata)
	helper.VectorStore.RemoveVector(collectionName, id)
}
