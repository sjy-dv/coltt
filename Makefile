PROTO_SRC_DIR=./proto/v1
PROTO_OUT_DIR=./proto/gen/v1
COMPILE_VERSION=V1

compile:
	@echo "Compiling proto files..."
	mkdir -p $(PROTO_OUT_DIR)
	find $(PROTO_SRC_DIR) -name "*.proto" | while read PROTO_FILE; do \
		FILE_NAME=$$(basename $$PROTO_FILE .proto); \
		mkdir -p "$(PROTO_OUT_DIR)/$${FILE_NAME}$(COMPILE_VERSION)"; \
		protoc -I$(PROTO_SRC_DIR) --gofast_out=plugins=grpc,paths=source_relative:"$(PROTO_OUT_DIR)/$${FILE_NAME}$(COMPILE_VERSION)" "$$PROTO_FILE"; \
	done

clean:
	@echo "Cleaning up generated files..."
	rm -rf $(PROTO_OUT_DIR)


add-license:
	- go-licenser -license ASL2 -licensor sjy-dv