init:
	@go mod tidy
	@go mod download

test: init
	@go test ./...

build: init
	@goreleaser release --clean --snapshot --skip publish

test-release: init
	@echo "==> Creating a temporary tag for release test..."
	@git tag -f v9.9.9-test
	@echo "==> Running goreleaser in release test mode (without publishing)..."
	@goreleaser release --skip publish --clean
	@echo "==> Cleaning up temporary tag..."
	@git tag -d v9.9.9-test
	@echo "==> Release test complete. Check the 'dist' directory for artifacts."