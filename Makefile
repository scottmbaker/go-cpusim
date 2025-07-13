all: build

.PHONY: build
build:
	go build -o build/_output/cpusim ./cmd

.PHONE: demo
demo: build
	./build/_output/cpusim -f roms/sbc-8251.rom

.PHONY: go-format
go-format:
	go fmt $(shell sh -c "go list ./...")

.PHONY: test
test:
	go test ./...

.PHONY: release
release:
	GOOS=linux GOARCH=amd64 go build -o release/linux/amd64/cpusim ./cmd
	GOOS=linux GOARCH=arm64 go build -o release/linux/arm64/cpusim ./cmd
	GOOS=windows GOARCH=amd64 go build -o release/windows/amd64/cpusim.exe ./cmd

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: clean
clean:
	rm -f release/linux/amd64/cpusim release/linux/arm64/cpusim release/windows/amd64/cpusim.exe build/_output/cpusim

