package widgets

import (
	"testing"

	"github.com/odvcencio/furry-ui/backend"
)

func TestDrawGaugeString(t *testing.T) {
	style := GaugeStyle{
		FillChar:   '█',
		EmptyChar:  '░',
		EmptyStyle: backend.DefaultStyle(),
	}

	tests := []struct {
		name   string
		width  int
		ratio  float64
		expect string
	}{
		{"empty", 10, 0.0, "░░░░░░░░░░"},
		{"full", 10, 1.0, "██████████"},
		{"half", 10, 0.5, "█████░░░░░"},
		{"quarter", 8, 0.25, "██░░░░░░"},
		{"over", 5, 1.5, "█████"},
		{"under", 5, -0.5, "░░░░░"},
		{"zero_width", 0, 0.5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DrawGaugeString(tt.width, tt.ratio, style)
			if got != tt.expect {
				t.Errorf("DrawGaugeString(%d, %.2f) = %q, want %q", tt.width, tt.ratio, got, tt.expect)
			}
		})
	}
}

func TestDrawGaugeSpans(t *testing.T) {
	green := backend.DefaultStyle().Foreground(backend.ColorGreen)
	amber := backend.DefaultStyle().Foreground(backend.ColorYellow)
	coral := backend.DefaultStyle().Foreground(backend.ColorRed)
	empty := backend.DefaultStyle()

	style := GaugeStyle{
		FillChar:  '█',
		EmptyChar: '░',
		Thresholds: []GaugeThreshold{
			{Ratio: 0.0, Style: green},
			{Ratio: 0.6, Style: amber},
			{Ratio: 0.85, Style: coral},
		},
		EmptyStyle: empty,
	}

	t.Run("gradient_colors", func(t *testing.T) {
		// At 100% fill with width 10:
		// cells 0-5 (ratios 0.0-0.5) = green
		// cells 6-7 (ratios 0.6-0.7) = amber
		// cells 8-9 (ratios 0.8-0.9) = coral
		spans := DrawGaugeSpans(10, 1.0, style)

		if len(spans) == 0 {
			t.Fatal("expected spans, got none")
		}

		// Should have multiple spans due to color changes
		totalRunes := 0
		for _, span := range spans {
			totalRunes += len([]rune(span.Text))
		}
		if totalRunes != 10 {
			t.Errorf("expected 10 total runes, got %d", totalRunes)
		}
	})

	t.Run("partial_fill", func(t *testing.T) {
		// At 50% fill with width 10, should have green fill + empty
		spans := DrawGaugeSpans(10, 0.5, style)

		if len(spans) < 2 {
			t.Fatalf("expected at least 2 spans (fill + empty), got %d", len(spans))
		}

		// Last span should be empty
		lastSpan := spans[len(spans)-1]
		if lastSpan.Style != empty {
			t.Error("expected last span to use empty style")
		}
	})

	t.Run("empty_gauge", func(t *testing.T) {
		spans := DrawGaugeSpans(10, 0.0, style)

		if len(spans) != 1 {
			t.Fatalf("expected 1 span for empty gauge, got %d", len(spans))
		}
		if spans[0].Style != empty {
			t.Error("expected empty style")
		}
		if spans[0].Text != "░░░░░░░░░░" {
			t.Errorf("expected all empty chars, got %q", spans[0].Text)
		}
	})
}

func TestStyleForRatio(t *testing.T) {
	green := backend.DefaultStyle().Foreground(backend.ColorGreen)
	amber := backend.DefaultStyle().Foreground(backend.ColorYellow)
	coral := backend.DefaultStyle().Foreground(backend.ColorRed)

	thresholds := []GaugeThreshold{
		{Ratio: 0.0, Style: green},
		{Ratio: 0.6, Style: amber},
		{Ratio: 0.85, Style: coral},
	}

	tests := []struct {
		ratio  float64
		expect backend.Style
	}{
		{0.0, green},
		{0.3, green},
		{0.59, green},
		{0.6, amber},
		{0.7, amber},
		{0.84, amber},
		{0.85, coral},
		{0.9, coral},
		{1.0, coral},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := styleForRatio(tt.ratio, thresholds)
			if got != tt.expect {
				t.Errorf("styleForRatio(%.2f) style mismatch", tt.ratio)
			}
		})
	}
}
