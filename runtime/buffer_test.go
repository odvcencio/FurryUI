package runtime

import (
	"testing"

	"github.com/odvcencio/furry-ui/backend"
)

func TestBuffer_New(t *testing.T) {
	b := NewBuffer(80, 24)

	w, h := b.Size()
	if w != 80 || h != 24 {
		t.Errorf("Size() = %d, %d; want 80, 24", w, h)
	}
}

func TestBuffer_SetGet(t *testing.T) {
	b := NewBuffer(10, 10)
	style := backend.DefaultStyle().Foreground(backend.ColorRGB(255, 0, 0))

	b.Set(5, 5, 'X', style)
	cell := b.Get(5, 5)

	if cell.Rune != 'X' {
		t.Errorf("Get() rune = %c, want X", cell.Rune)
	}
}

func TestBuffer_SetOutOfBounds(t *testing.T) {
	b := NewBuffer(10, 10)

	// Should not panic
	b.Set(-1, 5, 'X', backend.DefaultStyle())
	b.Set(100, 5, 'X', backend.DefaultStyle())
	b.Set(5, -1, 'X', backend.DefaultStyle())
	b.Set(5, 100, 'X', backend.DefaultStyle())

	// Get out of bounds returns space
	cell := b.Get(-1, -1)
	if cell.Rune != ' ' {
		t.Errorf("Get(-1,-1) = %c, want space", cell.Rune)
	}
}

func TestBuffer_SetString(t *testing.T) {
	b := NewBuffer(20, 5)
	style := backend.DefaultStyle()

	b.SetString(5, 2, "Hello", style)

	expected := []rune{'H', 'e', 'l', 'l', 'o'}
	for i, r := range expected {
		cell := b.Get(5+i, 2)
		if cell.Rune != r {
			t.Errorf("Get(%d, 2) = %c, want %c", 5+i, cell.Rune, r)
		}
	}
}

func TestBuffer_SetStringClips(t *testing.T) {
	b := NewBuffer(10, 5)
	style := backend.DefaultStyle()

	// String extends past buffer edge
	b.SetString(7, 2, "Hello", style)

	// Only first 3 chars should be written
	if b.Get(7, 2).Rune != 'H' {
		t.Error("expected H at (7,2)")
	}
	if b.Get(8, 2).Rune != 'e' {
		t.Error("expected e at (8,2)")
	}
	if b.Get(9, 2).Rune != 'l' {
		t.Error("expected l at (9,2)")
	}
}

func TestBuffer_Fill(t *testing.T) {
	b := NewBuffer(10, 10)
	style := backend.DefaultStyle()

	b.Fill(Rect{2, 2, 5, 5}, '#', style)

	// Check corners of filled region
	if b.Get(2, 2).Rune != '#' {
		t.Error("expected # at (2,2)")
	}
	if b.Get(6, 6).Rune != '#' {
		t.Error("expected # at (6,6)")
	}

	// Check outside filled region
	if b.Get(1, 1).Rune != 0 {
		t.Error("expected 0 at (1,1)")
	}
	if b.Get(7, 7).Rune != 0 {
		t.Error("expected 0 at (7,7)")
	}
}

func TestBuffer_Clear(t *testing.T) {
	b := NewBuffer(10, 10)
	style := backend.DefaultStyle()

	b.Fill(Rect{0, 0, 10, 10}, 'X', style)
	b.Clear()

	cell := b.Get(5, 5)
	if cell.Rune != ' ' {
		t.Errorf("After Clear(), cell = %c, want space", cell.Rune)
	}
}

func TestBuffer_Resize(t *testing.T) {
	b := NewBuffer(10, 10)
	style := backend.DefaultStyle()

	b.Set(5, 5, 'X', style)
	b.Resize(20, 20)

	w, h := b.Size()
	if w != 20 || h != 20 {
		t.Errorf("After Resize, Size() = %d, %d; want 20, 20", w, h)
	}

	// Original content preserved
	if b.Get(5, 5).Rune != 'X' {
		t.Error("content not preserved after resize")
	}
}

func TestBuffer_DrawBox(t *testing.T) {
	b := NewBuffer(10, 5)
	style := backend.DefaultStyle()

	b.DrawBox(Rect{0, 0, 10, 5}, style)

	// Check corners
	if b.Get(0, 0).Rune != '┌' {
		t.Errorf("top-left corner = %c, want ┌", b.Get(0, 0).Rune)
	}
	if b.Get(9, 0).Rune != '┐' {
		t.Errorf("top-right corner = %c, want ┐", b.Get(9, 0).Rune)
	}
	if b.Get(0, 4).Rune != '└' {
		t.Errorf("bottom-left corner = %c, want └", b.Get(0, 4).Rune)
	}
	if b.Get(9, 4).Rune != '┘' {
		t.Errorf("bottom-right corner = %c, want ┘", b.Get(9, 4).Rune)
	}

	// Check edges
	if b.Get(5, 0).Rune != '─' {
		t.Errorf("top edge = %c, want ─", b.Get(5, 0).Rune)
	}
	if b.Get(0, 2).Rune != '│' {
		t.Errorf("left edge = %c, want │", b.Get(0, 2).Rune)
	}
}

func TestSubBuffer(t *testing.T) {
	b := NewBuffer(20, 10)
	sub := b.Sub(Rect{5, 2, 10, 5})

	w, h := sub.Size()
	if w != 10 || h != 5 {
		t.Errorf("SubBuffer Size() = %d, %d; want 10, 5", w, h)
	}

	// Write to sub-buffer
	sub.Set(0, 0, 'X', backend.DefaultStyle())
	sub.SetString(1, 0, "Hello", backend.DefaultStyle())

	// Check it wrote to parent at offset
	if b.Get(5, 2).Rune != 'X' {
		t.Error("SubBuffer didn't write to parent at offset")
	}
	if b.Get(6, 2).Rune != 'H' {
		t.Error("SubBuffer SetString didn't work")
	}
}

func TestSubBuffer_Clips(t *testing.T) {
	b := NewBuffer(20, 10)
	sub := b.Sub(Rect{5, 2, 5, 3})

	// Should not write outside sub-buffer bounds
	sub.Set(10, 0, 'X', backend.DefaultStyle())
	sub.Set(-1, 0, 'Y', backend.DefaultStyle())

	// Parent should be unaffected outside sub region
	if b.Get(15, 2).Rune != 0 {
		t.Error("SubBuffer wrote outside bounds")
	}
	if b.Get(4, 2).Rune != 0 {
		t.Error("SubBuffer wrote at negative offset")
	}
}

func TestBuffer_DirtyTracking(t *testing.T) {
	b := NewBuffer(10, 10)

	// Initially not dirty
	if b.IsDirty() {
		t.Error("New buffer should not be dirty")
	}
	if b.DirtyCount() != 0 {
		t.Errorf("New buffer DirtyCount = %d, want 0", b.DirtyCount())
	}

	// Set a cell
	b.Set(5, 5, 'X', backend.DefaultStyle())

	if !b.IsDirty() {
		t.Error("Buffer should be dirty after Set")
	}
	if b.DirtyCount() != 1 {
		t.Errorf("DirtyCount = %d, want 1", b.DirtyCount())
	}
	if !b.IsCellDirty(5, 5) {
		t.Error("Cell (5,5) should be dirty")
	}
	if b.IsCellDirty(0, 0) {
		t.Error("Cell (0,0) should not be dirty")
	}
}

func TestBuffer_DirtyRect(t *testing.T) {
	b := NewBuffer(20, 20)

	// Set cells at different positions
	b.Set(5, 5, 'A', backend.DefaultStyle())
	b.Set(15, 15, 'B', backend.DefaultStyle())
	b.Set(10, 10, 'C', backend.DefaultStyle())

	dr := b.DirtyRect()

	// Dirty rect should encompass all dirty cells
	if dr.X != 5 {
		t.Errorf("DirtyRect X = %d, want 5", dr.X)
	}
	if dr.Y != 5 {
		t.Errorf("DirtyRect Y = %d, want 5", dr.Y)
	}
	if dr.Width != 11 { // 15 - 5 + 1 = 11
		t.Errorf("DirtyRect Width = %d, want 11", dr.Width)
	}
	if dr.Height != 11 { // 15 - 5 + 1 = 11
		t.Errorf("DirtyRect Height = %d, want 11", dr.Height)
	}
}

func TestBuffer_ClearDirty(t *testing.T) {
	b := NewBuffer(10, 10)
	b.Set(5, 5, 'X', backend.DefaultStyle())

	if !b.IsDirty() {
		t.Fatal("Buffer should be dirty")
	}

	b.ClearDirty()

	if b.IsDirty() {
		t.Error("Buffer should not be dirty after ClearDirty")
	}
	if b.DirtyCount() != 0 {
		t.Errorf("DirtyCount = %d after ClearDirty, want 0", b.DirtyCount())
	}
	if b.DirtyRect() != (Rect{}) {
		t.Errorf("DirtyRect = %v after ClearDirty, want empty", b.DirtyRect())
	}
}

func TestBuffer_MarkAllDirty(t *testing.T) {
	b := NewBuffer(10, 10)

	b.MarkAllDirty()

	if !b.IsDirty() {
		t.Error("Buffer should be dirty after MarkAllDirty")
	}
	if b.DirtyCount() != 100 {
		t.Errorf("DirtyCount = %d, want 100", b.DirtyCount())
	}
	dr := b.DirtyRect()
	if dr.Width != 10 || dr.Height != 10 {
		t.Errorf("DirtyRect = %v, want full buffer", dr)
	}
}

func TestBuffer_ForEachDirtyCell(t *testing.T) {
	b := NewBuffer(10, 10)
	style := backend.DefaultStyle()

	b.Set(2, 2, 'A', style)
	b.Set(5, 5, 'B', style)
	b.Set(8, 8, 'C', style)

	visited := make(map[string]rune)
	b.ForEachDirtyCell(func(x, y int, cell Cell) {
		key := string(rune('0'+x)) + string(rune('0'+y))
		visited[key] = cell.Rune
	})

	if len(visited) != 3 {
		t.Errorf("ForEachDirtyCell visited %d cells, want 3", len(visited))
	}
	if visited["22"] != 'A' {
		t.Error("Should have visited (2,2) with 'A'")
	}
	if visited["55"] != 'B' {
		t.Error("Should have visited (5,5) with 'B'")
	}
	if visited["88"] != 'C' {
		t.Error("Should have visited (8,8) with 'C'")
	}
}

func TestBuffer_ForEachDirtyCell_Empty(t *testing.T) {
	b := NewBuffer(10, 10)

	count := 0
	b.ForEachDirtyCell(func(x, y int, cell Cell) {
		count++
	})

	if count != 0 {
		t.Errorf("ForEachDirtyCell visited %d cells on empty buffer, want 0", count)
	}
}

func TestBuffer_ForEachDirtyCell_MostDirty(t *testing.T) {
	b := NewBuffer(10, 10)

	// Mark more than half the cells dirty
	for y := 0; y < 10; y++ {
		for x := 0; x < 6; x++ {
			b.Set(x, y, 'X', backend.DefaultStyle())
		}
	}

	count := 0
	b.ForEachDirtyCell(func(x, y int, cell Cell) {
		count++
	})

	if count != 60 {
		t.Errorf("ForEachDirtyCell visited %d cells, want 60", count)
	}
}

func TestBuffer_IsCellDirty_OutOfBounds(t *testing.T) {
	b := NewBuffer(10, 10)

	// Out of bounds should return false
	if b.IsCellDirty(-1, 5) {
		t.Error("IsCellDirty(-1,5) should be false")
	}
	if b.IsCellDirty(10, 5) {
		t.Error("IsCellDirty(10,5) should be false")
	}
	if b.IsCellDirty(5, -1) {
		t.Error("IsCellDirty(5,-1) should be false")
	}
	if b.IsCellDirty(5, 10) {
		t.Error("IsCellDirty(5,10) should be false")
	}
}

func TestBuffer_SetString_OutOfBounds(t *testing.T) {
	b := NewBuffer(10, 5)
	style := backend.DefaultStyle()

	// Y out of bounds - should not panic
	b.SetString(0, -1, "Hello", style)
	b.SetString(0, 10, "Hello", style)

	// Partial X out of bounds - should write what fits
	b.SetString(-2, 0, "Hello", style)
	// First two chars are clipped, so "llo" should start at x=0
	if b.Get(0, 0).Rune != 'l' {
		t.Errorf("Expected 'l' at (0,0), got '%c'", b.Get(0, 0).Rune)
	}
}

func TestBuffer_ResizeSmaller(t *testing.T) {
	b := NewBuffer(20, 20)
	style := backend.DefaultStyle()

	b.Set(5, 5, 'X', style)
	b.Set(15, 15, 'Y', style)

	b.Resize(10, 10)

	w, h := b.Size()
	if w != 10 || h != 10 {
		t.Errorf("After Resize, Size() = %d, %d; want 10, 10", w, h)
	}

	// Original content within new bounds preserved
	if b.Get(5, 5).Rune != 'X' {
		t.Error("content at (5,5) not preserved after resize")
	}
}

func TestBuffer_ResizeSameSize(t *testing.T) {
	b := NewBuffer(10, 10)
	style := backend.DefaultStyle()

	b.Set(5, 5, 'X', style)
	b.ClearDirty()

	b.Resize(10, 10) // Same size

	// Should be no-op, not dirty
	if b.IsDirty() {
		t.Error("Resize to same size should not mark dirty")
	}
}

func TestBuffer_DrawRoundedBox(t *testing.T) {
	b := NewBuffer(10, 5)
	style := backend.DefaultStyle()

	b.DrawRoundedBox(Rect{0, 0, 10, 5}, style)

	// Check rounded corners
	if b.Get(0, 0).Rune != 0x256D { // ╭
		t.Errorf("top-left corner = %c, want ╭", b.Get(0, 0).Rune)
	}
	if b.Get(9, 0).Rune != 0x256E { // ╮
		t.Errorf("top-right corner = %c, want ╮", b.Get(9, 0).Rune)
	}
	if b.Get(0, 4).Rune != 0x2570 { // ╰
		t.Errorf("bottom-left corner = %c, want ╰", b.Get(0, 4).Rune)
	}
	if b.Get(9, 4).Rune != 0x256F { // ╯
		t.Errorf("bottom-right corner = %c, want ╯", b.Get(9, 4).Rune)
	}
}

func TestBuffer_DrawBoxSmall(t *testing.T) {
	b := NewBuffer(10, 10)
	style := backend.DefaultStyle()

	// Box too small - should not draw
	b.DrawBox(Rect{0, 0, 1, 1}, style)
	b.DrawRoundedBox(Rect{0, 0, 1, 1}, style)

	// No panic and no drawing
	if b.Get(0, 0).Rune != 0 {
		t.Error("DrawBox should not draw for small rects")
	}
}

func TestBuffer_ClearRect(t *testing.T) {
	b := NewBuffer(20, 20)
	style := backend.DefaultStyle()

	// Fill entire buffer
	b.Fill(Rect{0, 0, 20, 20}, 'X', style)

	// Clear a region
	b.ClearRect(Rect{5, 5, 10, 10})

	// Check cleared region
	if b.Get(10, 10).Rune != ' ' {
		t.Error("ClearRect should fill with spaces")
	}

	// Check outside cleared region
	if b.Get(0, 0).Rune != 'X' {
		t.Error("ClearRect should not affect outside region")
	}
}

func TestBuffer_SetSameValue(t *testing.T) {
	b := NewBuffer(10, 10)
	style := backend.DefaultStyle()

	b.Set(5, 5, 'X', style)
	b.ClearDirty()

	// Set same value - should not mark dirty
	b.Set(5, 5, 'X', style)

	if b.IsDirty() {
		t.Error("Setting same value should not mark cell dirty")
	}
}

func TestSubBuffer_SetStringClips(t *testing.T) {
	b := NewBuffer(20, 10)
	sub := b.Sub(Rect{5, 2, 5, 3})

	// String that extends beyond sub-buffer width
	sub.SetString(3, 0, "Hello", backend.DefaultStyle())

	// Only first 2 chars fit (positions 3 and 4)
	if b.Get(8, 2).Rune != 'H' {
		t.Errorf("Expected 'H' at (8,2), got '%c'", b.Get(8, 2).Rune)
	}
	if b.Get(9, 2).Rune != 'e' {
		t.Errorf("Expected 'e' at (9,2), got '%c'", b.Get(9, 2).Rune)
	}
	// Position 10 is outside sub-buffer
	if b.Get(10, 2).Rune != 0 {
		t.Error("SubBuffer should not write beyond its bounds")
	}
}

func TestSubBuffer_SetStringYOutOfBounds(t *testing.T) {
	b := NewBuffer(20, 10)
	sub := b.Sub(Rect{5, 2, 5, 3})

	// Should not panic
	sub.SetString(0, -1, "Hello", backend.DefaultStyle())
	sub.SetString(0, 10, "Hello", backend.DefaultStyle())
}

func TestSubBuffer_Fill(t *testing.T) {
	b := NewBuffer(20, 10)
	sub := b.Sub(Rect{5, 2, 5, 3})

	sub.Fill(Rect{0, 0, 5, 3}, '#', backend.DefaultStyle())

	// Check filled region in parent
	if b.Get(5, 2).Rune != '#' {
		t.Error("SubBuffer Fill should write to parent")
	}
	if b.Get(9, 4).Rune != '#' {
		t.Error("SubBuffer Fill should cover entire region")
	}

	// Check outside sub-buffer
	if b.Get(4, 2).Rune != 0 {
		t.Error("SubBuffer Fill should not write outside bounds")
	}
}

func TestSubBuffer_FillClipsToSubBounds(t *testing.T) {
	b := NewBuffer(20, 10)
	sub := b.Sub(Rect{5, 2, 5, 3})

	// Fill region that extends beyond sub-buffer
	sub.Fill(Rect{3, 1, 10, 10}, '#', backend.DefaultStyle())

	// Should only fill within sub-buffer bounds
	if b.Get(10, 3).Rune != 0 {
		t.Error("SubBuffer Fill should clip to sub-buffer bounds")
	}
}

func TestSubBuffer_FillNoIntersection(t *testing.T) {
	b := NewBuffer(20, 10)
	sub := b.Sub(Rect{5, 2, 5, 3})

	// Fill region completely outside sub-buffer
	sub.Fill(Rect{100, 100, 5, 5}, '#', backend.DefaultStyle())

	// Nothing should be written
	if b.IsDirty() {
		// Check if any cells were actually modified
		count := 0
		b.ForEachDirtyCell(func(x, y int, cell Cell) {
			count++
		})
		// If dirty count is 0, that's fine
		if count > 0 {
			t.Error("SubBuffer Fill with no intersection should not modify buffer")
		}
	}
}

func TestSubBuffer_Clear(t *testing.T) {
	b := NewBuffer(20, 10)
	sub := b.Sub(Rect{5, 2, 5, 3})

	// First fill with something
	sub.Fill(Rect{0, 0, 5, 3}, 'X', backend.DefaultStyle())

	// Then clear
	sub.Clear()

	if b.Get(5, 2).Rune != ' ' {
		t.Error("SubBuffer Clear should fill with spaces")
	}
}
