
app_name=dotenv-myvault
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
	@pwsh ./build.ps1 -targetOS $(target_os) -path $(target_folder)

test: init
	@go test

build: init 
	@echo "building for $(OS) -> $(target_path)"
	@go build -o $(target_path) main.go