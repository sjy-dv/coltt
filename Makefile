GOGOPROTO_SRC_DIR=./idl/gogoproto/v1
GOGOPROTO_OUT_DIR=./gen/gogoprotoc/v1
PROTO_SRC_DIR=./idl/proto/v2
PROTO_OUT_DIR=./gen/protoc/v2
COMPILE_VERSION=V2

docker-build-bin:
	- env GOOS=linux go build -o ./bin/ cmd/root/main.go

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



clean:
	@echo "Cleaning up generated files..."
	rm -rf $(PROTO_OUT_DIR)


add-license:
	- go-licenser -license ASL2 -licensor sjy-dv

simple-docker:
	- docker build --file simple.dockerfile -t nnv:simple .
	- docker run nnv:simple -p 50051:50051 -d

simple2-docker:
	- docker build --file simple.v2.dockerfile -t nnv:cgo .

test:
	- go test -v --count=1 ./pkg/sharding
# - go test -v --count=1 ./storage
	- go test -v --count=1 ./pkg/flat
# - go test -v --count=1 ./pkg/hnsw

compress-float:
	- go test -v --count=1 ./pkg/compresshelper