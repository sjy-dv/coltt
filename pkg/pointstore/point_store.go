package pointstore

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/sjy-dv/nnv/pkg/conversion"
	"github.com/sjy-dv/nnv/pkg/models"
	"github.com/sjy-dv/nnv/storage"
)

const POINTS_STORAGE_NAME = "points"

var ErrPointDoesNotExist = errors.New("point does not exist")

type ShardPoint struct {
	models.Point
	NodeId uint64
}

func PointKey(id uuid.UUID, suffix byte) []byte {
	key := [18]byte{}
	key[0] = 'p'
	copy(key[1:], id[:])
	key[17] = suffix
	return key[:]
}

func SetPoint(storage storage.Storage, point ShardPoint) error {
	// ---------------------------
	// Set matching ids
	if err := storage.Put(conversion.NodeKey(point.NodeId, 'i'), point.Id[:]); err != nil {
		return fmt.Errorf("could not set point id: %w", err)
	}
	if err := storage.Put(PointKey(point.Id, 'i'), conversion.Uint64ToBytes(point.NodeId)); err != nil {
		return fmt.Errorf("could not set node id: %w", err)
	}
	// ---------------------------
	// Handle point data
	if len(point.Data) > 0 {
		if err := storage.Put(conversion.NodeKey(point.NodeId, 'd'), point.Data); err != nil {
			return fmt.Errorf("could not set point data: %w", err)
		}
	} else {
		if err := storage.Delete(conversion.NodeKey(point.NodeId, 'd')); err != nil {
			return fmt.Errorf("could not delete empty point data: %w", err)
		}
	}
	return nil
}

func CheckPointExists(storage storage.Storage, pointId uuid.UUID) (bool, error) {
	v := storage.Get(PointKey(pointId, 'i'))
	return v != nil, nil
}

func GetPointNodeIdByUUID(storage storage.Storage, pointId uuid.UUID) (uint64, error) {
	nodeIdBytes := storage.Get(PointKey(pointId, 'i'))
	if nodeIdBytes == nil {
		return 0, ErrPointDoesNotExist
	}
	nodeId := conversion.BytesToUint64(nodeIdBytes)
	return nodeId, nil
}

func GetPointByUUID(storage storage.Storage, pointId uuid.UUID) (ShardPoint, error) {
	nodeId, err := GetPointNodeIdByUUID(storage, pointId)
	if err != nil {
		return ShardPoint{}, err
	}
	data := storage.Get(conversion.NodeKey(nodeId, 'd'))
	sp := ShardPoint{
		Point: models.Point{
			Id:   pointId,
			Data: data,
		},
		NodeId: nodeId,
	}
	return sp, nil
}

func GetPointByNodeId(storage storage.Storage, nodeId uint64, withData bool) (ShardPoint, error) {
	pointIdBytes := storage.Get(conversion.NodeKey(nodeId, 'i'))
	if pointIdBytes == nil {
		return ShardPoint{}, ErrPointDoesNotExist
	}
	pointId, err := uuid.FromBytes(pointIdBytes)
	if err != nil {
		return ShardPoint{}, fmt.Errorf("could not parse point id: %w", err)
	}
	var data []byte
	if withData {
		data = storage.Get(conversion.NodeKey(nodeId, 'd'))
	}
	sp := ShardPoint{
		Point: models.Point{
			Id:   pointId,
			Data: data,
		},
		NodeId: nodeId,
	}
	return sp, nil
}

func DeletePoint(storage storage.Storage, pointId uuid.UUID, nodeId uint64) error {
	if err := storage.Delete(PointKey(pointId, 'i')); err != nil {
		return fmt.Errorf("could not delete point id: %w", err)
	}
	if err := storage.Delete(conversion.NodeKey(nodeId, 'i')); err != nil {
		return fmt.Errorf("could not delete node id: %w", err)
	}
	if err := storage.Delete(conversion.NodeKey(nodeId, 'd')); err != nil {
		return fmt.Errorf("could not delete point data: %w", err)
	}
	return nil
}
