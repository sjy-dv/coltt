package highmem

import (
	"context"
	"fmt"

	"github.com/sjy-dv/nnv/gen/protoc/v2/dataCoordinatorV2"
	"github.com/vmihailenco/msgpack/v5"
)

func (xx *HighMem) Insert(
	ctx context.Context,
	req *dataCoordinatorV2.ModifyDataset,
) (*dataCoordinatorV2.Response, error) {

	type reply struct {
		Result *dataCoordinatorV2.Response
		Error  error
	}

	c := make(chan reply, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				c <- reply{
					Result: nil,
					Error:  fmt.Errorf(UncaughtPanicError, r),
				}
			}
		}()
		metadata := make(map[string]interface{})
		err := msgpack.Unmarshal(req.GetMetadata(), &metadata)
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV2.Response{
					Status: false,
					Error: &dataCoordinatorV2.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		autoId := autoCommitID()
		// first add data
		xx.Collections[req.GetCollectionName()].collectionLock.Lock()
		xx.Collections[req.GetCollectionName()].Data[autoId] = metadata
		xx.Collections[req.GetCollectionName()].collectionLock.Unlock()
		//second build index
		err = indexdb.indexes[req.GetCollectionName()].Add(autoId, metadata)
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV2.Response{
					Status: false,
					Error: &dataCoordinatorV2.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		//last build vector index
		err = tensorLinker.tensors[req.GetCollectionName()].Add(autoId, req.GetVector())
		if err != nil {
			c <- reply{
				Result: &dataCoordinatorV2.Response{
					Status: false,
					Error: &dataCoordinatorV2.Error{
						ErrorMessage: err.Error(),
						ErrorCode:    dataCoordinatorV2.ErrorCode_INTERNAL_FUNC_ERROR,
					},
				},
			}
			return
		}
		c <- reply{
			Result: &dataCoordinatorV2.Response{
				Status: true,
			},
		}
	}()
	res := <-c
	return res.Result, res.Error
}
