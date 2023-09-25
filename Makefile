BINARY_NAME=PaymentGoApp
.DEFAULT_GOAL := run-linux

update:
	go mod download

build-windows: update
	set GOARCH=amd64
	set GOOS=windows
	go build -o ./build/${BINARY_NAME}-windows.exe ./cmd/main.go
build-linux: update
	export GOARCH=amd64
	export GOOS=linux
	go build -o ./build/${BINARY_NAME}-linux ./cmd/main.go
build-debian: update
	export GOARCH=amd64
	export GOOS=debian
	go build -o ./build/${BINARY_NAME}-debian ./cmd/main.go

run-windows: build-windows
	"./build/${BINARY_NAME}-windows.exe"

run-linux: build-linux
	./build/${BINARY_NAME}-linux

run-debian: build-debian
	./build/${BINARY_NAME}-debian

