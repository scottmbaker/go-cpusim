all: build

.PHONY: build
build: build8008 build4004

.PHONY: build8008
build8008:
	go build -o build/_output/cpusim8008 ./cmd/cpusim8008

.PHONY: build4004
build4004:
	go build -o build/_output/cpusim4004 ./cmd/cpusim4004

.PHONE: demo
demo: build
	./build/_output/cpusim8008 -f roms/sbc-8251.rom

.PHONE: demo4004
demo4004: build
	./build/_output/cpusim4004 -f roms/scott-4004-uart.rom

.PHONY: go-format
go-format:
	go fmt $(shell sh -c "go list ./...")

.PHONY: test
test:
	go test ./...

.PHONY: release
release:
	GOOS=linux GOARCH=amd64 go build -o release/linux/amd64/cpusim8008 ./cmd/cpusim8008
	GOOS=linux GOARCH=arm64 go build -o release/linux/arm64/cpusim8008 ./cmd/cpusim8008
	GOOS=windows GOARCH=amd64 go build -o release/windows/amd64/cpusim8008.exe ./cmd/cpusim8008

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: clean
clean:
	rm -rf release
	rm -f build/_output/cpusim*
