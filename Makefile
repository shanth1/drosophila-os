.DEFAULT_GOAL := build

.PHONY: plugins ui-install ui-dev ui-build ui-run build run test check clean wasm-lab snn-lab pipeline-lab

ui-install:
	npm --prefix ui ci

ui-dev: ui-install
	npm --prefix ui run dev

ui-build: ui-install
	npm --prefix ui run build

ui-run: build
	./bin/drosophila -mode ui $(ARGS)

plugins:
	mkdir -p plugins/compiled
	GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -mod=readonly -buildmode=c-shared -o plugins/compiled/sensor_random.wasm ./plugins/src/sensor_random
	GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -mod=readonly -buildmode=c-shared -o plugins/compiled/sensor_fixed.wasm ./plugins/src/sensor_fixed

build: plugins ui-build
	mkdir -p bin
	CGO_ENABLED=0 go build -mod=readonly -o bin/drosophila ./cmd/drosophila

run: build
	./bin/drosophila $(ARGS)

test: plugins ui-build
	CGO_ENABLED=0 go test -mod=readonly ./...

check: test
	CGO_ENABLED=0 go vet -mod=readonly ./...

clean:
	rm -f bin/drosophila plugins/compiled/sensor_random.wasm plugins/compiled/sensor_fixed.wasm

wasm-lab:
	CGO_ENABLED=0 go test -mod=readonly -v -count=1 ./examples/wasm

snn-lab:
	CGO_ENABLED=0 go test -mod=readonly -v -count=1 ./examples/snn

pipeline-lab:
	CGO_ENABLED=0 go test -mod=readonly -v -count=1 ./examples/pipeline
