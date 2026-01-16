package theme

import (
	"testing"

	"github.com/odvcencio/furry-ui/compositor"
)

func TestDefaultTheme(t *testing.T) {
	th := DefaultTheme()

	if th == nil {
		t.Fatal("DefaultTheme() returned nil")
	}

	// Verify core palette is set
	if th.Background == (compositor.Style{}) {
		t.Error("Background style not set")
	}
	if th.Surface == (compositor.Style{}) {
		t.Error("Surface style not set")
	}
	if th.SurfaceRaised == (compositor.Style{}) {
		t.Error("SurfaceRaised style not set")
	}

	// Verify text hierarchy
	if th.TextPrimary == (compositor.Style{}) {
		t.Error("TextPrimary style not set")
	}
	if th.TextSecondary == (compositor.Style{}) {
		t.Error("TextSecondary style not set")
	}
	if th.TextMuted == (compositor.Style{}) {
		t.Error("TextMuted style not set")
	}

	// Verify accent colors
	if th.Accent == (compositor.Style{}) {
		t.Error("Accent style not set")
	}
	if th.AccentGlow == (compositor.Style{}) {
		t.Error("AccentGlow style not set")
	}
	if th.ElectricBlue == (compositor.Style{}) {
		t.Error("ElectricBlue style not set")
	}
	if th.Coral == (compositor.Style{}) {
		t.Error("Coral style not set")
	}
	if th.Teal == (compositor.Style{}) {
		t.Error("Teal style not set")
	}
	if th.BlueGlow == (compositor.Style{}) {
		t.Error("BlueGlow style not set")
	}
	if th.PurpleGlow == (compositor.Style{}) {
		t.Error("PurpleGlow style not set")
	}
	if th.CoralGlow == (compositor.Style{}) {
		t.Error("CoralGlow style not set")
	}

	// Verify semantic colors
	if th.Success == (compositor.Style{}) {
		t.Error("Success style not set")
	}
	if th.Warning == (compositor.Style{}) {
		t.Error("Warning style not set")
	}
	if th.Error == (compositor.Style{}) {
		t.Error("Error style not set")
	}
	if th.Info == (compositor.Style{}) {
		t.Error("Info style not set")
	}

	// Verify message source styles
	if th.User == (compositor.Style{}) {
		t.Error("User style not set")
	}
	if th.Assistant == (compositor.Style{}) {
		t.Error("Assistant style not set")
	}
	if th.System == (compositor.Style{}) {
		t.Error("System style not set")
	}
	if th.Tool == (compositor.Style{}) {
		t.Error("Tool style not set")
	}

	// Verify UI element styles
	if th.Border == (compositor.Style{}) {
		t.Error("Border style not set")
	}
	if th.BorderFocus == (compositor.Style{}) {
		t.Error("BorderFocus style not set")
	}
	if th.Selection == (compositor.Style{}) {
		t.Error("Selection style not set")
	}

	// Verify mode indicators
	if th.ModeNormal == (compositor.Style{}) {
		t.Error("ModeNormal style not set")
	}
	if th.ModeShell == (compositor.Style{}) {
		t.Error("ModeShell style not set")
	}
	if th.ModeEnv == (compositor.Style{}) {
		t.Error("ModeEnv style not set")
	}
	if th.ModeSearch == (compositor.Style{}) {
		t.Error("ModeSearch style not set")
	}
}

func TestSymbols(t *testing.T) {
	// Test bullets and markers
	if Symbols.Bullet == "" {
		t.Error("Bullet symbol not set")
	}
	if Symbols.Check == "" {
		t.Error("Check symbol not set")
	}
	if Symbols.Cross == "" {
		t.Error("Cross symbol not set")
	}

	// Test border characters
	if Symbols.BorderTopLeft == "" {
		t.Error("BorderTopLeft not set")
	}
	if Symbols.BorderTopRight == "" {
		t.Error("BorderTopRight not set")
	}
	if Symbols.BorderBottomLeft == "" {
		t.Error("BorderBottomLeft not set")
	}
	if Symbols.BorderBottomRight == "" {
		t.Error("BorderBottomRight not set")
	}
	if Symbols.BorderHorizontal == "" {
		t.Error("BorderHorizontal not set")
	}
	if Symbols.BorderVertical == "" {
		t.Error("BorderVertical not set")
	}

	// Test spinner frames
	if len(Symbols.Spinner) == 0 {
		t.Error("Spinner frames not set")
	}

	// Test message prefixes
	if Symbols.User == "" {
		t.Error("User prefix not set")
	}
	if Symbols.Assistant == "" {
		t.Error("Assistant prefix not set")
	}
	if Symbols.System == "" {
		t.Error("System prefix not set")
	}
	if Symbols.Tool == "" {
		t.Error("Tool prefix not set")
	}

	// Test mode indicators
	if Symbols.ModeNormal == "" {
		t.Error("ModeNormal indicator not set")
	}
	if Symbols.ModeShell == "" {
		t.Error("ModeShell indicator not set")
	}
	if Symbols.ModeEnv == "" {
		t.Error("ModeEnv indicator not set")
	}
	if Symbols.ModeSearch == "" {
		t.Error("ModeSearch indicator not set")
	}
}

func TestLayout(t *testing.T) {
	// Test padding values
	if Layout.PaddingXS <= 0 {
		t.Error("PaddingXS should be positive")
	}
	if Layout.PaddingSM <= 0 {
		t.Error("PaddingSM should be positive")
	}
	if Layout.PaddingMD <= 0 {
		t.Error("PaddingMD should be positive")
	}

	// Test component dimensions
	if Layout.HeaderHeight <= 0 {
		t.Error("HeaderHeight should be positive")
	}
	if Layout.StatusHeight <= 0 {
		t.Error("StatusHeight should be positive")
	}
	if Layout.InputMinHeight <= 0 {
		t.Error("InputMinHeight should be positive")
	}
	if Layout.PickerMaxHeight <= 0 {
		t.Error("PickerMaxHeight should be positive")
	}
	if Layout.ScrollbarWidth <= 0 {
		t.Error("ScrollbarWidth should be positive")
	}
}

func TestThemeColorContrast(t *testing.T) {
	th := DefaultTheme()

	// Verify accent color has visible foreground
	if th.Accent.FG.Mode == compositor.ColorModeDefault {
		t.Error("Accent should have explicit foreground color")
	}

	// Verify AccentGlow is bold
	if !th.AccentGlow.Bold {
		t.Error("AccentGlow should be bold")
	}

	// Verify System style is italic
	if !th.System.Italic {
		t.Error("System style should be italic")
	}

	// Verify Thinking style is italic
	if !th.Thinking.Italic {
		t.Error("Thinking style should be italic")
	}

	// Verify mode styles that should be bold
	if !th.ModeShell.Bold {
		t.Error("ModeShell should be bold")
	}
	if !th.ModeEnv.Bold {
		t.Error("ModeEnv should be bold")
	}
	if !th.ModeSearch.Bold {
		t.Error("ModeSearch should be bold")
	}
}

func BenchmarkDefaultTheme(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultTheme()
	}
}
