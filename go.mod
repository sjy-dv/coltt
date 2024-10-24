module github.com/sjy-dv/nnv

go 1.23.1

require (
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/google/uuid v1.6.0
	golang.org/x/sync v0.8.0
	google.golang.org/grpc v1.67.1
	google.golang.org/protobuf v1.35.1
)

require (
	github.com/bits-and-blooms/bitset v1.12.0 // indirect
	github.com/dgraph-io/ristretto v1.0.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/klauspost/compress v1.17.10 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mschoch/smat v0.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
)

require (
	github.com/RoaringBitmap/roaring v1.9.4
	github.com/dgraph-io/badger/v4 v4.3.1
	github.com/gofrs/flock v0.12.1
	github.com/klauspost/cpuid v1.3.1
	github.com/rs/zerolog v1.33.0
	github.com/shirou/gopsutil v3.21.11+incompatible
	go.etcd.io/bbolt v1.3.8
	golang.org/x/crypto v0.28.0
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241015192408-796eee8c2d53 // indirect
)

replace go.etcd.io/bbolt => github.com/yanxiaoqi932/bbolt v1.3.9-0.20240829105042-5b817c5f51f8
