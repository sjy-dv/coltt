GOGOPROTO_SRC_DIR=./idl/gogoproto/v1
GOGOPROTO_OUT_DIR=./gen/gogoprotoc/v1
PROTO_SRC_DIR=./idl/proto/v1
PROTO_OUT_DIR=./gen/protoc/v1
COMPILE_VERSION=V1

compile-gogoproto:
	@echo "Compiling proto files..."
	mkdir -p $(GOGOPROTO_OUT_DIR)
	find $(GOGOPROTO_SRC_DIR) -name "*.proto" | while read PROTO_FILE; do \
		FILE_NAME=$$(basename $$PROTO_FILE .proto); \
		mkdir -p "$(GOGOPROTO_OUT_DIR)/$${FILE_NAME}$(COMPILE_VERSION)"; \
		protoc -I$(GOGOPROTO_SRC_DIR) --gofast_out=plugins=grpc,paths=source_relative:"$(GOGOPROTO_OUT_DIR)/$${FILE_NAME}$(COMPILE_VERSION)" "$$PROTO_FILE"; \
	done

compile-proto:
	- protoc --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=$(PROTO_OUT_DIR) \
	--proto_path=./ $(PROTO_SRC_DIR)/*.proto && protoc --go_out=$(PROTO_OUT_DIR) --proto_path=./ $(PROTO_SRC_DIR)/*.proto

clean:
	@echo "Cleaning up generated files..."
	rm -rf $(PROTO_OUT_DIR)


add-license:
	- go-licenser -license ASL2 -licensor sjy-dv