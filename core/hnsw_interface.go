package core

import (
	"github.com/sjy-dv/nnv/pkg/gomath"
	"github.com/sjy-dv/nnv/pkg/models"
	"github.com/sjy-dv/nnv/pkg/queue"
)

type HNSWIndex interface {
	CreateCollection(collectionName string, config models.HnswConfig, params models.ProductQuantizerParameters) error
	Genesis(collectionName string, config models.HnswConfig)
	Insert(collectionName string, commitID uint64, vector gomath.Vector) error
	Delete(collectionName string, nodeId uint64) error
	Search(collectionName string, vector gomath.Vector,
		topCandidates *queue.PriorityQueue, topK int, efSearch int) error
	FailAppointNode(collectionName string, failID uint64) error
	AppointNode(collectionName string) uint64
	IsEmpty(collectionName string) bool
}
