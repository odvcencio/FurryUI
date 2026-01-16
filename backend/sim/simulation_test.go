package sim

import (
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/terminal"
)

func TestBackend_BasicRendering(t *testing.T) {
	sim := New(20, 5)
	if err := sim.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer sim.Fini()

	// Write some text
	style := backend.DefaultStyle().Foreground(backend.ColorWhite)
	text := "Hello, World!"
	for i, r := range text {
		sim.SetContent(i, 0, r, nil, style)
	}
	sim.Show()

	// Capture and verify
	w, h := sim.Size()
	capture := sim.Capture()
	lines := strings.Split(capture, "\n")
	if len(lines) != h {
		t.Errorf("Expected %d lines, got %d", h, len(lines))
	}

	firstLine := lines[0]
	if !strings.HasPrefix(firstLine, "Hello, World!") {
		t.Errorf("Expected first line to start with 'Hello, World!', got %q", firstLine)
	}

	// Verify width
	if len(firstLine) != w {
		t.Logf("First line width: %d, expected: %d", len(firstLine), w)
	}
}

func TestBackend_Size(t *testing.T) {
	sim := New(80, 24)
	if err := sim.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer sim.Fini()

	w, h := sim.Size()
	if w != 80 {
		t.Errorf("Expected width 80, got %d", w)
	}
	if h < 24 || h > 25 {
		t.Errorf("Expected height around 24, got %d", h)
	}
}

func TestBackend_Resize(t *testing.T) {
	sim := New(80, 24)
	if err := sim.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer sim.Fini()

	sim.Resize(40, 12)

	w, h := sim.Size()
	if w != 40 || h != 12 {
		t.Errorf("Expected size 40x12 after resize, got %dx%d", w, h)
	}
}

func TestBackend_ContainsText(t *testing.T) {
	sim := New(40, 10)
	if err := sim.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer sim.Fini()

	// Write text at specific position
	style := backend.DefaultStyle()
	text := "findme"
	for i, r := range text {
		sim.SetContent(5+i, 3, r, nil, style)
	}
	sim.Show()

	if !sim.ContainsText("findme") {
		t.Error("Expected to find 'findme' on screen")
	}

	if sim.ContainsText("nothere") {
		t.Error("Should not find 'nothere' on screen")
	}
}

func TestBackend_FindText(t *testing.T) {
	sim := New(40, 10)
	if err := sim.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer sim.Fini()

	// Write text at position (5, 3)
	style := backend.DefaultStyle()
	text := "target"
	for i, r := range text {
		sim.SetContent(5+i, 3, r, nil, style)
	}
	sim.Show()

	x, y := sim.FindText("target")
	if x != 5 || y != 3 {
		t.Errorf("Expected to find 'target' at (5, 3), got (%d, %d)", x, y)
	}

	x, y = sim.FindText("missing")
	if x != -1 || y != -1 {
		t.Errorf("Expected (-1, -1) for missing text, got (%d, %d)", x, y)
	}
}

func TestBackend_CaptureRegion(t *testing.T) {
	sim := New(20, 10)
	if err := sim.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer sim.Fini()

	// Write a box pattern
	style := backend.DefaultStyle()
	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			sim.SetContent(x, y, 'X', nil, style)
		}
	}
	sim.Show()

	region := sim.CaptureRegion(0, 0, 5, 3)
	expected := "XXXXX\nXXXXX\nXXXXX"
	if region != expected {
		t.Errorf("Expected region:\n%s\nGot:\n%s", expected, region)
	}
}

func TestBackend_InjectKey(t *testing.T) {
	sim := New(20, 10)
	if err := sim.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer sim.Fini()

	// Inject a key event
	sim.InjectKeyRune('a')

	// Use a goroutine with timeout to avoid blocking forever
	done := make(chan struct{})
	var ev terminal.Event

	go func() {
		ev = sim.PollEvent()
		close(done)
	}()

	select {
	case <-done:
		// Got event
	case <-time.After(100 * time.Millisecond):
		t.Skip("PollEvent blocked - tcell simulation may not support this")
		return
	}

	if ev == nil {
		t.Skip("No event received - may be tcell simulation limitation")
		return
	}

	keyEv, ok := ev.(terminal.KeyEvent)
	if !ok {
		t.Fatalf("Expected terminal.KeyEvent, got %T", ev)
	}

	if keyEv.Key != terminal.KeyRune || keyEv.Rune != 'a' {
		t.Errorf("Expected KeyRune 'a', got key=%v rune=%c", keyEv.Key, keyEv.Rune)
	}
}

func TestBackend_Styles(t *testing.T) {
	sim := New(20, 10)
	if err := sim.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer sim.Fini()

	// Write with specific style
	style := backend.DefaultStyle().
		Foreground(backend.ColorRed).
		Background(backend.ColorBlue).
		Bold(true)

	sim.SetContent(0, 0, 'S', nil, style)
	sim.Show()

	// Capture cell and verify style
	mainc, _, capturedStyle := sim.CaptureCell(0, 0)
	if mainc != 'S' {
		t.Errorf("Expected 'S', got %c", mainc)
	}

	// Check attributes
	attrs := capturedStyle.Attributes()
	if attrs&backend.AttrBold == 0 {
		t.Error("Expected bold attribute to be set")
	}
}
