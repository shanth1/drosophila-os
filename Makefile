.DEFAULT_GOAL := build

.PHONY: plugins build run test check clean wasm-lab snn-lab pipeline-lab

plugins:
	mkdir -p plugins/compiled
	GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -mod=readonly -buildmode=c-shared -o plugins/compiled/sensor_random.wasm ./plugins/src/sensor_random

build: plugins
	mkdir -p bin
	CGO_ENABLED=0 go build -mod=readonly -o bin/drosophila ./cmd/drosophila

run: build
	./bin/drosophila $(ARGS)

test: plugins
	CGO_ENABLED=0 go test -mod=readonly ./...

check: test
	CGO_ENABLED=0 go vet -mod=readonly ./...

clean:
	rm -f bin/drosophila plugins/compiled/sensor_random.wasm

wasm-lab:
	CGO_ENABLED=0 go test -mod=readonly -v -count=1 ./examples/wasm

snn-lab:
	CGO_ENABLED=0 go test -mod=readonly -v -count=1 ./examples/snn

pipeline-lab:
	CGO_ENABLED=0 go test -mod=readonly -v -count=1 ./examples/pipeline
