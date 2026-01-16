// Package theme provides a unified visual design system for terminal UIs.
// Inspired by Dark Elegance: rich blacks, subtle depth, glowing accents.
package theme

import (
	"github.com/odvcencio/furry-ui/compositor"
)

// Theme defines the complete visual language for the TUI.
type Theme struct {
	// Core palette
	Background    compositor.Style // Primary canvas
	Surface       compositor.Style // Elevated surfaces (cards, panels)
	SurfaceRaised compositor.Style // Higher elevation
	SurfaceDim    compositor.Style // Recessed areas

	// Text hierarchy
	TextPrimary   compositor.Style // Main content
	TextSecondary compositor.Style // Supporting text
	TextMuted     compositor.Style // Hints, placeholders
	TextInverse   compositor.Style // Text on accent backgrounds

	// Accent colors
	Accent       compositor.Style // Primary action, highlights
	AccentDim    compositor.Style // Subtle accent usage
	AccentGlow   compositor.Style // Emphasis, active states
	ElectricBlue compositor.Style // Active processes, streaming
	Coral        compositor.Style // Warnings, errors, attention
	Teal         compositor.Style // Informational, metrics

	// Glow variants
	BlueGlow   compositor.Style
	PurpleGlow compositor.Style
	CoralGlow  compositor.Style

	// Semantic colors
	Success compositor.Style
	Warning compositor.Style
	Error   compositor.Style
	Info    compositor.Style

	// Message sources
	User      compositor.Style
	Assistant compositor.Style
	System    compositor.Style
	Tool      compositor.Style
	Thinking  compositor.Style

	// UI elements
	Border      compositor.Style
	BorderFocus compositor.Style
	Selection   compositor.Style
	SearchMatch compositor.Style
	Scrollbar   compositor.Style
	ScrollThumb compositor.Style

	// Mode indicators
	ModeNormal compositor.Style
	ModeShell  compositor.Style
	ModeEnv    compositor.Style
	ModeSearch compositor.Style

	// Special
	Logo    compositor.Style
	Spinner compositor.Style
}

// DefaultTheme returns the Dark Elegance theme.
func DefaultTheme() *Theme {
	return &Theme{
		// Core palette - deep blacks with subtle blue undertone
		Background:    compositor.DefaultStyle().WithBG(compositor.RGB(12, 12, 16)),
		Surface:       compositor.DefaultStyle().WithBG(compositor.RGB(22, 22, 28)),
		SurfaceRaised: compositor.DefaultStyle().WithBG(compositor.RGB(32, 32, 40)),
		SurfaceDim:    compositor.DefaultStyle().WithBG(compositor.RGB(8, 8, 10)),

		// Text hierarchy - warm whites
		TextPrimary:   compositor.DefaultStyle().WithFG(compositor.RGB(240, 238, 232)),
		TextSecondary: compositor.DefaultStyle().WithFG(compositor.RGB(160, 158, 150)),
		TextMuted:     compositor.DefaultStyle().WithFG(compositor.RGB(100, 98, 92)),
		TextInverse:   compositor.DefaultStyle().WithFG(compositor.RGB(12, 12, 16)),

		// Accent - warm amber/gold (memorable, warm, inviting)
		Accent:       compositor.DefaultStyle().WithFG(compositor.RGB(255, 183, 77)),
		AccentDim:    compositor.DefaultStyle().WithFG(compositor.RGB(180, 130, 60)),
		AccentGlow:   compositor.DefaultStyle().WithFG(compositor.RGB(255, 200, 100)).WithBold(true),
		ElectricBlue: compositor.DefaultStyle().WithFG(compositor.RGB(79, 195, 247)),
		Coral:        compositor.DefaultStyle().WithFG(compositor.RGB(255, 138, 101)),
		Teal:         compositor.DefaultStyle().WithFG(compositor.RGB(77, 182, 172)),

		// Glow variants - softened highlights
		BlueGlow:   compositor.DefaultStyle().WithFG(compositor.RGB(120, 210, 255)).WithDim(true),
		PurpleGlow: compositor.DefaultStyle().WithFG(compositor.RGB(210, 160, 255)).WithDim(true),
		CoralGlow:  compositor.DefaultStyle().WithFG(compositor.RGB(255, 170, 150)).WithDim(true),

		// Semantic colors
		Success: compositor.DefaultStyle().WithFG(compositor.RGB(134, 239, 172)),
		Warning: compositor.DefaultStyle().WithFG(compositor.RGB(255, 138, 101)),
		Error:   compositor.DefaultStyle().WithFG(compositor.RGB(255, 110, 90)),
		Info:    compositor.DefaultStyle().WithFG(compositor.RGB(77, 182, 172)),

		// Message sources - each has distinct character
		User:      compositor.DefaultStyle().WithFG(compositor.RGB(134, 239, 172)), // Fresh green
		Assistant: compositor.DefaultStyle().WithFG(compositor.RGB(255, 183, 77)),  // Warm amber
		System:    compositor.DefaultStyle().WithFG(compositor.RGB(160, 158, 150)).WithItalic(true),
		Tool:      compositor.DefaultStyle().WithFG(compositor.RGB(192, 132, 252)), // Purple
		Thinking:  compositor.DefaultStyle().WithFG(compositor.RGB(100, 98, 92)).WithItalic(true),

		// UI elements
		Border:      compositor.DefaultStyle().WithFG(compositor.RGB(50, 50, 60)),
		BorderFocus: compositor.DefaultStyle().WithFG(compositor.RGB(255, 183, 77)),
		Selection:   compositor.DefaultStyle().WithBG(compositor.RGB(60, 60, 80)),
		SearchMatch: compositor.DefaultStyle().WithBG(compositor.RGB(120, 90, 20)).WithFG(compositor.RGB(255, 255, 255)),
		Scrollbar:   compositor.DefaultStyle().WithFG(compositor.RGB(50, 50, 60)),
		ScrollThumb: compositor.DefaultStyle().WithFG(compositor.RGB(100, 100, 110)),

		// Mode indicators
		ModeNormal: compositor.DefaultStyle().WithFG(compositor.RGB(160, 158, 150)),
		ModeShell:  compositor.DefaultStyle().WithFG(compositor.RGB(134, 239, 172)).WithBold(true),
		ModeEnv:    compositor.DefaultStyle().WithFG(compositor.RGB(147, 197, 253)).WithBold(true),
		ModeSearch: compositor.DefaultStyle().WithFG(compositor.RGB(253, 224, 71)).WithBold(true),

		// Special
		Logo:    compositor.DefaultStyle().WithFG(compositor.RGB(255, 183, 77)).WithBold(true),
		Spinner: compositor.DefaultStyle().WithFG(compositor.RGB(255, 183, 77)),
	}
}

// Symbols provides consistent iconography.
var Symbols = struct {
	// Bullets and markers
	Bullet      string
	BulletEmpty string
	Arrow       string
	ArrowRight  string
	ArrowDown   string
	Check       string
	Cross       string
	Dot         string
	Ring        string

	// Borders (rounded)
	BorderTopLeft     string
	BorderTopRight    string
	BorderBottomLeft  string
	BorderBottomRight string
	BorderHorizontal  string
	BorderVertical    string

	// UI elements
	Spinner      []string
	Progress     string
	ProgressFill string
	Scrollbar    string
	ScrollThumb  string

	// Message prefixes
	User      string
	Assistant string
	System    string
	Tool      string
	Thinking  string

	// Mode indicators
	ModeNormal string
	ModeShell  string
	ModeEnv    string
	ModeSearch string

	// File types
	FileDefault  string
	FileFolder   string
	FileGo       string
	FileJS       string
	FileTS       string
	FilePython   string
	FileMarkdown string
	FileYAML     string
	FileJSON     string
}{
	// Bullets and markers
	Bullet:      "●",
	BulletEmpty: "○",
	Arrow:       "›",
	ArrowRight:  "→",
	ArrowDown:   "↓",
	Check:       "✓",
	Cross:       "✗",
	Dot:         "·",
	Ring:        "◌",

	// Borders (rounded for softer feel)
	BorderTopLeft:     "╭",
	BorderTopRight:    "╮",
	BorderBottomLeft:  "╰",
	BorderBottomRight: "╯",
	BorderHorizontal:  "─",
	BorderVertical:    "│",

	// UI elements
	Spinner:      []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	Progress:     "░",
	ProgressFill: "█",
	Scrollbar:    "░",
	ScrollThumb:  "█",

	// Message prefixes
	User:      "❯",
	Assistant: "●",
	System:    "◆",
	Tool:      "◇",
	Thinking:  "…",

	// Mode indicators
	ModeNormal: "λ",
	ModeShell:  "!",
	ModeEnv:    "$",
	ModeSearch: "/",

	// File types
	FileDefault:  "◇",
	FileFolder:   "▸",
	FileGo:       "◈",
	FileJS:       "◆",
	FileTS:       "◆",
	FilePython:   "◈",
	FileMarkdown: "◇",
	FileYAML:     "◇",
	FileJSON:     "◇",
}

// Layout defines standard spacing and dimensions.
var Layout = struct {
	// Padding
	PaddingXS int
	PaddingSM int
	PaddingMD int
	PaddingLG int
	PaddingXL int

	// Margins
	MarginXS int
	MarginSM int
	MarginMD int
	MarginLG int
	MarginXL int

	// Component dimensions
	HeaderHeight    int
	StatusHeight    int
	InputMinHeight  int
	PickerMaxHeight int
	ScrollbarWidth  int
}{
	// Padding
	PaddingXS: 1,
	PaddingSM: 2,
	PaddingMD: 3,
	PaddingLG: 4,
	PaddingXL: 6,

	// Margins
	MarginXS: 1,
	MarginSM: 2,
	MarginMD: 3,
	MarginLG: 4,
	MarginXL: 6,

	// Component dimensions
	HeaderHeight:    1,
	StatusHeight:    1,
	InputMinHeight:  3,
	PickerMaxHeight: 15,
	ScrollbarWidth:  1,
}
