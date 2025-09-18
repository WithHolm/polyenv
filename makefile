init:
	@go mod tidy
	@go mod download

test: init
	@go test ./internal/... ./cmd/... -race
# -coverprofile=.\.local\coverage.out
# 	@go tool cover -html=.\.local\coverage.out -o .\.local\coverage.html

check:
	@echo "==> Running golangci-lint..."
	@golangci-lint run
	@echo "==> Running staticcheck..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@staticcheck ./...
	@echo "==> Running govulncheck..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...


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