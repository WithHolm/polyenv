
app_name=polyenv
target_folder=./build
target_os=windows-amd64,linux-amd64,linux-ppc64le,darwin-amd64,darwin-arm64
target_path=$(target_folder)/$(app_name)

ifeq ($(OS),Windows_NT)
	target_path=$(target_folder)/$(app_name).exe
endif


init:
	@go mod tidy
	@go mod download

build-ci: init
	@pwsh ./.github/script/build.ps1 -targetOS $(target_os) -path $(target_folder)

test: init
	@go test ./...

build: init
	@goreleaser release --clean --snapshot
