GOGOPROTO_SRC_DIR=./idl/gogoproto/v1
GOGOPROTO_OUT_DIR=./gen/gogoprotoc/v1
PROTO_SRC_DIR=./idl/proto/v4
PROTO_OUT_DIR=./gen/protoc/v4
COMPILE_VERSION=V4

start:
	- go run cmd/root/main.go -mode=root

start-edge:
	- go run cmd/root/main.go -mode=edge

start-rc:
	- go run cmd/root/main.go -mode=experimental

docker-build-bin:
	- env GOOS=linux go build -o ./bin/ cmd/root/main.go

make-minio:
	- docker run -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin \
  -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"



compile-gogoproto:
	@echo "Compiling proto files..."
	mkdir -p $(GOGOPROTO_OUT_DIR)
	find $(GOGOPROTO_SRC_DIR) -name "*.proto" | while read PROTO_FILE; do \
		FILE_NAME=$$(basename $$PROTO_FILE .proto); \
		mkdir -p "$(GOGOPROTO_OUT_DIR)/$${FILE_NAME}$(COMPILE_VERSION)"; \
		protoc -I$(GOGOPROTO_SRC_DIR) --gofast_out=plugins=grpc,paths=source_relative:"$(GOGOPROTO_OUT_DIR)/$${FILE_NAME}$(COMPILE_VERSION)" "$$PROTO_FILE"; \
	done

compile-proto-py:
	mkdir -p $(PROTO_OUT_DIR)
	python -m grpc_tools.protoc -I./idl/proto/v2 --proto_path=./ --python_out=./gen/protoc/v2 --pyi_out=./gen/protoc/v2 --grpc_python_out=./gen/protoc/v2 ./idl/proto/v2/*.proto

compile-proto:
	- protoc --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=$(PROTO_OUT_DIR) \
	--proto_path=./ $(PROTO_SRC_DIR)/*.proto && protoc --go_out=$(PROTO_OUT_DIR) --proto_path=./ $(PROTO_SRC_DIR)/*.proto

clean:
	@echo "Cleaning up generated files..."
	rm -rf $(PROTO_OUT_DIR)


add-license:
	- go-licenser -license ASL2 -licensor sjy-dv

edge-docker:
	- docker build --file edge.dockerfile -t coltt:edge .
	

simple-docker:
	- docker build --file simple.dockerfile -t coltt:simple .
	- docker run coltt:simple -p 50051:50051 -d

simple2-docker:
	- docker build --file simple.v2.dockerfile -t coltt:cgo .

test:
	- go test -v --count=1 ./pkg/sharding
# - go test -v --count=1 ./storage
# - go test -v --count=1 ./pkg/flat
# - go test -v --count=1 ./pkg/hnsw

compress-float:
	- go test -v --count=1 ./pkg/compresshelper


bench-milvus-boot:
	- docker-compose -f ./benchmark/milvus.docker.compose.yaml -p benchmark-milvus up

e2e-test:
	@echo "e2e Test [HNSW Commit & Load]"
	- go run e2e/e2e_hnsw.go