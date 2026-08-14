package render

// Color — a CSS color; the canvas adapter applies it as-is.
type Color string

const (
	ColorRunning  Color = "#22c55e" // green
	ColorBlocked  Color = "#ef4444" // red
	ColorFinished Color = "#9ca3af" // gray
	ColorChannel  Color = "#3b82f6" // blue — channel link
)

type OpKind int

const (
	OpLine OpKind = iota
	OpText
)

// Op — one atomic draw command (IR).
type Op struct {
	Text           string
	Color          Color
	X1, Y1, X2, Y2 float64
	Kind           OpKind
}
