package backend

// RowWriter is an optional optimization for bulk row updates.
type RowWriter interface {
	SetRow(y int, startX int, cells []Cell)
}
