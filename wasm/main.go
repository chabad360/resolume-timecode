package main

import (
	//"fmt"
	. "github.com/siongui/godom/wasm"
)

func main() {
	testdiv := Document.QuerySelector("#testdiv")
	testdiv.Set("innerHTML", "-00:00:00.000")
	addr := "localhost"

	startOSC(&testdiv, addr)
}
