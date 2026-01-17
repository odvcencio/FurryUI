package backend

// RectWriter is an optional optimization for bulk rectangle updates.
// The cells slice is row-major and must have width*height entries.
type RectWriter interface {
	SetRect(x, y, width, height int, cells []Cell)
}
