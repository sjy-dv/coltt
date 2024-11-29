package core

import "github.com/sjy-dv/nnv/pkg/snowflake"

var autogen *snowflake.Node

func autoCommitID() uint64 {
	x := autogen.Generate()
	if x.Int64() < 0 {
		return uint64(-x.Int64())
	}
	return uint64(x.Int64())
}

func NewIdGenerator() error {
	gen, err := snowflake.NewNode(0)
	if err != nil {
		return err
	}
	autogen = gen
	return nil
}
