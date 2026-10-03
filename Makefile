.PHONY: plugins run test check wasm-lab

plugins:
	mkdir -p plugins/compiled
	GOOS=wasip1 GOARCH=wasm CGO_ENABLED=0 go build -mod=readonly -buildmode=c-shared -o plugins/compiled/sensor_random.wasm ./plugins/src/sensor_random

run: plugins
	CGO_ENABLED=0 go run -mod=readonly ./cmd/drosophila $(ARGS)

test: plugins
	CGO_ENABLED=0 go test -mod=readonly ./...

check: test
	CGO_ENABLED=0 go vet -mod=readonly ./...

wasm-lab:
	CGO_ENABLED=0 go test -mod=readonly -v -count=1 -run '^TestWASMFunctionCall$$' ./examples/wasm
