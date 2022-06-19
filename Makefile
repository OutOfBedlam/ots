all: ots

.PHONY: ots
ots:
	@./scripts/build.sh ots

regen:
	@echo "protoc regen tiles..."
	@cd tiles && \
		protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		tiles.proto

package:
	@scripts/package.sh ots Linux   linux   amd64
	@scripts/package.sh ots Mac     darwin  arm64
	@scripts/package.sh ots Mac     darwin  amd64


test:
	@go test \
		./geom \
		./glob \
		./logging \
		./projection \
		./tiles