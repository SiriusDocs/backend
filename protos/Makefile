PROTO_PATH := proto

.PHONY: gen
gen: clean
	@protoc -I $(PROTO_PATH) \
		--go_out=./gen/go --go_opt=paths=source_relative \
		--go-grpc_out=./gen/go --go-grpc_opt=paths=source_relative \
		$(shell find $(PROTO_PATH) -name "*.proto")
	@echo "Generation completed!"

.PHONY: clean
clean:
	rm -rf gen/go/*

