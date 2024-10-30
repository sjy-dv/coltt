package matchdbgo

import (
	"strconv"

	"github.com/sjy-dv/nnv/kv"
)

var mdb *kv.DB

func Open() error {
	opts := kv.DefaultOptions
	opts.DirPath = "./data_dir/matchid"
	db, err := kv.Open(opts)
	if err != nil {
		return err
	}
	mdb = db
	return nil
}

func Close() error {
	return mdb.Close()
}

func Get(key string) (uint32, error) {

	nodeId, err := mdb.Get([]byte(key))
	if err != nil {
		return 0, err
	}
	uintId, err := strconv.ParseUint(string(nodeId), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(uintId), nil
}

func Set(key string, val uint32) error {
	bk := []byte(key)
	bv := []byte(strconv.FormatUint(uint64(val), 10))
	return mdb.Put(bk, bv)
}

func Delete(key string) error {
	return mdb.Delete([]byte(key))
}
