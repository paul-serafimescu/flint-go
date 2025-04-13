.PHONY: proto clean check setup-hooks

default: proto setup-hooks

proto:
	@echo "Generating code from .proto files..."
	mkdir -p pkg/api
	protoc --proto_path=./ \
	       --go_out=pkg --go_opt=paths=source_relative \
	       --go-grpc_out=pkg --go-grpc_opt=paths=source_relative \
	       api/compute.proto

	@echo "Done."

clean:
	@echo "Cleaning generated files..."
	rm -rf pkg/api/*.pb.go
	@echo "Done."

check:
	@echo "Verifying if .pb.go files are up to date..."
	@TEMP_DIR=$$(mktemp -d) && \
	mkdir -p $$TEMP_DIR/pkg/api && \
	protoc \
	  --proto_path=./ \
	  --go_out=pkg --go_opt=paths=source_relative \
	  --go-grpc_out=pkg --go-grpc_opt=paths=source_relative \
	  api/compute.proto \
	  --go_out=$$TEMP_DIR/pkg --go_opt=paths=source_relative \
	  --go-grpc_out=$$TEMP_DIR/pkg --go-grpc_opt=paths=source_relative && \
	diff -qr $$TEMP_DIR/pkg/api pkg/api || (echo "Out of date files. Run \`make proto\`." && exit 1)
	@echo "Up to date."


setup-hooks:
	@echo "Configuring Git to use shared hooks in .githooks/"
	@git config core.hooksPath .githooks
	@chmod +x .githooks/* || true
	@echo "Done."

