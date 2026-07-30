package render

// Color — CSS-цвет; canvas-адаптер применит его как есть.
type Color string

const (
	ColorRunning  Color = "#22c55e" // зелёный
	ColorBlocked  Color = "#ef4444" // красный
	ColorFinished Color = "#9ca3af" // серый
	ColorChannel  Color = "#3b82f6" // синий — связь канала
)

type OpKind int

const (
	OpLine OpKind = iota
	OpText
)

// Op — одна атомарная команда отрисовки (IR).
type Op struct {
	Text           string
	Color          Color
	X1, Y1, X2, Y2 float64
	Kind           OpKind
}
