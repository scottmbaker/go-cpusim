all: build

.PHONY: build
build: build8008 build4004 build4004-bigram buildz80

.PHONY: build8008
build8008:
	go build -o build/_output/cpusim8008 ./cmd/cpusim8008

.PHONY: build4004
build4004:
	go build -o build/_output/cpusim4004 ./cmd/cpusim4004

.PHONY: build4004-bigram
build4004-bigram:
	go build -o build/_output/cpusim4004-bigram ./cmd/cpusim4004-bigram

.PHONY: buildz80
buildz80:
	go build -o build/_output/cpusim-z80-rc2014 ./cmd/cpusim-z80-rc2014

.PHONY: testdata-z80
testdata-z80:
	mkdir -p pkg/cpusim/cpuz80/testdata
	cd pkg/cpusim/cpuz80/testdata && git clone --depth 1 https://github.com/SingleStepTests/z80 .

.PHONY: demo
demo: build
	./build/_output/cpusim8008 -f roms/sbc-8251.rom

.PHONY: demo4004
demo4004: build
	./build/_output/cpusim4004 -f roms/scott-4004-uart.rom

.PHONY: demo-z80-rc2014
demo-z80-rc2014:
	./build/_output/cpusim-z80-rc2014 -f roms/z80/nostos512k.rom --cf-image disks/z80/nostos-cf.img --cf-offset 1024

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
	GOOS=linux GOARCH=amd64 go build -o release/linux/amd64/cpusimz80 ./cmd/cpusimz80
	GOOS=linux GOARCH=arm64 go build -o release/linux/arm64/cpusimz80 ./cmd/cpusimz80
	GOOS=windows GOARCH=amd64 go build -o release/windows/amd64/cpusimz80.exe ./cmd/cpusimz80

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: clean
clean:
	rm -rf release
	rm -f build/_output/cpusim*
