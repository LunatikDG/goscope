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
	el := js.Global().Get("document").Call("getElementById", id)

	dpr := js.Global().Get("devicePixelRatio").Float()
	if dpr == 0 {
		dpr = 1
	}
	// логический размер берём из CSS-раскладки (clientWidth), с фолбэком
	cssW := el.Get("clientWidth").Float()
	cssH := el.Get("clientHeight").Float()
	if cssW == 0 {
		cssW = 640
	}
	if cssH == 0 {
		cssH = 360
	}
	// backing store — в dpr раз крупнее (важно: сброс width обнуляет трансформацию)
	el.Set("width", int(cssW*dpr))
	el.Set("height", int(cssH*dpr))

	ctx := el.Call("getContext", "2d")
	ctx.Call("scale", dpr, dpr) // теперь рисуем в CSS-пикселях, резко

	return canvasCtx{ctx: ctx, width: cssW, height: cssH}
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
