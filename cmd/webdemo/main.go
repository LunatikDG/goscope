//go:build js && wasm

package main

import (
	"syscall/js"
)

func main() {
	ctx := js.Global().Get("document").Call("getElementById", "canvas").Call("getContext", "2d")
	ctx.Set("fillStyle", "#22c55e")
	ctx.Call("fillRect", 20, 20, 150, 80)

	select {}
}