package runtime

import (
	"testing"

	"github.com/odvcencio/furry-ui/backend"
)

// BenchmarkBuffer_Set measures single cell writes.
func BenchmarkBuffer_Set(b *testing.B) {
	buf := NewBuffer(120, 40)
	style := backend.DefaultStyle()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := i % 120
		y := (i / 120) % 40
		buf.Set(x, y, 'A', style)
	}
}

// BenchmarkBuffer_SetString measures string writes.
func BenchmarkBuffer_SetString(b *testing.B) {
	buf := NewBuffer(120, 40)
	style := backend.DefaultStyle()
	text := "Hello, World!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		y := i % 40
		buf.SetString(0, y, text, style)
	}
}

// BenchmarkBuffer_Fill measures rectangle fills.
func BenchmarkBuffer_Fill(b *testing.B) {
	buf := NewBuffer(120, 40)
	style := backend.DefaultStyle()
	rect := Rect{X: 10, Y: 5, Width: 50, Height: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Fill(rect, ' ', style)
	}
}

// BenchmarkBuffer_Clear measures full clear.
func BenchmarkBuffer_Clear(b *testing.B) {
	buf := NewBuffer(120, 40)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear()
	}
}

// BenchmarkBuffer_ForEachDirtyCell_Few measures iteration with few dirty cells.
func BenchmarkBuffer_ForEachDirtyCell_Few(b *testing.B) {
	buf := NewBuffer(120, 40)
	style := backend.DefaultStyle()

	// Mark a few cells dirty
	for i := 0; i < 50; i++ {
		buf.Set(i, 0, 'X', style)
	}

	var count int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count = 0
		buf.ForEachDirtyCell(func(x, y int, cell Cell) {
			count++
		})
	}
	_ = count
}

// BenchmarkBuffer_ForEachDirtyCell_Many measures iteration with many dirty cells.
func BenchmarkBuffer_ForEachDirtyCell_Many(b *testing.B) {
	buf := NewBuffer(120, 40)
	buf.MarkAllDirty()

	var count int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count = 0
		buf.ForEachDirtyCell(func(x, y int, cell Cell) {
			count++
		})
	}
	_ = count
}

// BenchmarkBuffer_ClearDirty measures dirty flag clearing.
func BenchmarkBuffer_ClearDirty(b *testing.B) {
	buf := NewBuffer(120, 40)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.MarkAllDirty()
		buf.ClearDirty()
	}
}

// BenchmarkBuffer_Resize measures resize operations.
func BenchmarkBuffer_Resize(b *testing.B) {
	buf := NewBuffer(120, 40)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			buf.Resize(130, 45)
		} else {
			buf.Resize(120, 40)
		}
	}
}

// BenchmarkBuffer_DrawBox measures box drawing.
func BenchmarkBuffer_DrawBox(b *testing.B) {
	buf := NewBuffer(120, 40)
	style := backend.DefaultStyle()
	rect := Rect{X: 5, Y: 5, Width: 60, Height: 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.DrawBox(rect, style)
	}
}

// BenchmarkSubBuffer_SetString measures sub-buffer string writes.
func BenchmarkSubBuffer_SetString(b *testing.B) {
	buf := NewBuffer(120, 40)
	sub := buf.Sub(Rect{X: 10, Y: 10, Width: 80, Height: 20})
	style := backend.DefaultStyle()
	text := "This is a test message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		y := i % 20
		sub.SetString(0, y, text, style)
	}
}

// BenchmarkBuffer_RenderCycle simulates a typical render cycle.
func BenchmarkBuffer_RenderCycle(b *testing.B) {
	buf := NewBuffer(120, 40)
	style := backend.DefaultStyle()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate widget rendering
		buf.Fill(Rect{0, 0, 120, 1}, ' ', style)  // Header
		buf.SetString(2, 0, "Sample", style)

		buf.Fill(Rect{0, 1, 120, 38}, ' ', style) // Main area
		for line := 0; line < 10; line++ {
			buf.SetString(2, 2+line, "This is a sample chat message line", style)
		}

		buf.Fill(Rect{0, 39, 120, 1}, ' ', style) // Status
		buf.SetString(2, 39, "Ready", style)

		// Clear dirty for next frame
		buf.ClearDirty()
	}
}
