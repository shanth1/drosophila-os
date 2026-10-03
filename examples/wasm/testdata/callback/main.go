//go:build wasip1 && wasm

package main

//go:wasmimport host report
func report(value int32)

//go:wasmexport run
func run() {
	report(42)
}

func main() {}
