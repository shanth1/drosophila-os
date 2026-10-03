package plugins

import (
	_ "embed"
)

//go:embed compiled/sensor_random.wasm
var SensorRandomWASM []byte

//go:embed compiled/sensor_fixed.wasm
var SensorFixedWASM []byte
