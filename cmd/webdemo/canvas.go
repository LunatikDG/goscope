//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/LunatikDG/goscope/internal/render"
)

// canvasCtx — тонкая обёртка над 2D-контекстом canvas.
type canvasCtx struct {
	ctx           js.Value
	width, height float64
}

func newCanvas(id string) canvasCtx {
	canvas := js.Global().Get("document").Call("getElementById", id)
	return canvasCtx{
		ctx:    canvas.Call("getContext", "2d"),
		width:  canvas.Get("width").Float(),
		height: canvas.Get("height").Float(),
	}
}

func (c canvasCtx) clear() {
	c.ctx.Call("clearRect", 0, 0, c.width, c.height)
}

// draw исполняет список команд отрисовки.
func (c canvasCtx) draw(ops []render.Op) {
	for _, op := range ops {
		switch op.Kind {
		case render.OpLine:
			c.ctx.Set("strokeStyle", string(op.Color))
			c.ctx.Set("lineWidth", 4)
			c.ctx.Call("beginPath")
			c.ctx.Call("moveTo", op.X1, op.Y1)
			c.ctx.Call("lineTo", op.X2, op.Y2)
			c.ctx.Call("stroke")
		case render.OpText:
			c.ctx.Set("fillStyle", string(op.Color))
			c.ctx.Set("font", "12px sans-serif")
			c.ctx.Call("fillText", op.Text, op.X1, op.Y1)
		}
	}
}
