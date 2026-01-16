package widgets

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/runtime"
)

var updateSnapshots = flag.Bool("update-snapshots", false, "Update golden snapshot files")

// renderToString renders a widget to a string for snapshot comparison.
func renderToString(w runtime.Widget, width, height int) string {
	// Create buffer
	buf := runtime.NewBuffer(width, height)

	// Measure and layout
	constraints := runtime.Constraints{MaxWidth: width, MaxHeight: height}
	w.Measure(constraints)
	w.Layout(runtime.Rect{X: 0, Y: 0, Width: width, Height: height})

	// Render
	ctx := runtime.RenderContext{Buffer: buf}
	w.Render(ctx)

	// Convert to string
	var sb strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			cell := buf.Get(x, y)
			if cell.Rune == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(cell.Rune)
			}
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}

// assertSnapshot compares rendered output against a golden file.
func assertSnapshot(t *testing.T, name string, actual string) {
	t.Helper()

	goldenPath := filepath.Join("testdata", name+".golden")

	if *updateSnapshots {
		// Update mode: write the actual output
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("failed to create testdata dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(actual), 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("Updated snapshot: %s", goldenPath)
		return
	}

	// Compare mode: read expected and compare
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("snapshot file not found: %s\nRun with -update-snapshots to create it.\nActual output:\n%s", goldenPath, actual)
		}
		t.Fatalf("failed to read golden file: %v", err)
	}

	if actual != string(expected) {
		t.Errorf("snapshot mismatch for %s\n\nExpected:\n%s\n\nActual:\n%s\n\nRun with -update-snapshots to update.", name, string(expected), actual)
	}
}

// TestSnapshot_Palette tests the palette widget rendering.
func TestSnapshot_Palette(t *testing.T) {
	p := NewPaletteWidget("Commands")
	p.SetItems([]PaletteItem{
		{ID: "1", Category: "Session", Label: "New Conversation", Shortcut: "/new"},
		{ID: "2", Category: "Session", Label: "Clear Messages", Shortcut: "/clear"},
		{ID: "3", Category: "View", Label: "Toggle Sidebar", Shortcut: "Ctrl+B"},
	})
	p.SetStyles(
		backend.DefaultStyle(),
		backend.DefaultStyle(),
		backend.DefaultStyle().Bold(true),
		backend.DefaultStyle(),
		backend.DefaultStyle(),
		backend.DefaultStyle().Reverse(true),
		backend.DefaultStyle(),
	)
	p.Focus()

	output := renderToString(p, 60, 12)
	assertSnapshot(t, "palette", output)
}
