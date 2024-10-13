package coordinator

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/sjy-dv/vemoo/cluster"
)

type Coordinator struct {
	config *Config

	db   *badger.DB
	conn *cluster.Conn
}
