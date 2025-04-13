.PHONY: proto clean check setup-hooks

proto:
	@echo "🛠 Generating Go code from .proto files..."
	mkdir -p pkg/api
	protoc --proto_path=./ \
	       --go_out=pkg --go_opt=paths=source_relative \
	       --go-grpc_out=pkg --go-grpc_opt=paths=source_relative \
	       api/compute.proto

	@echo "✅ Done."

clean:
	@echo "🧹 Cleaning generated files..."
	rm -rf pkg/api/*.pb.go
	@echo "✅ Clean complete."

check:
	@echo "🔍 Verifying if .pb.go files are up to date..."
	@TEMP_DIR=$$(mktemp -d) && \
	cp -r api $$TEMP_DIR && \
	cd $$TEMP_DIR && \
	protoc --proto_path=api --go_out=go_out --go-grpc_out=go_out api/*.proto && \
	diff -r $$TEMP_DIR/go_out ../generated || (echo "❌ Out of date files. Run \`make proto\`." && exit 1)
	@echo "✅ Up to date."

setup-hooks:
	@echo "Configuring Git to use shared hooks in .githooks/"
	@git config core.hooksPath .githooks
	@chmod +x .githooks/* || true
	@echo "Done."

