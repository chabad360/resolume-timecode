package main

import (
	"syscall/js"
)

func main() {
	testdiv := js.Global().Get("document").Call("getElementById", "testdiv")
	testdiv.Set("innerHTML", "-00:00:00.000")
	addr := "localhost"

	startOSC(&testdiv, addr)
}
