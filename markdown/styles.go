// Package markdown provides markdown parsing and rendering for terminal UIs.
// It uses goldmark for parsing and custom rendering to the compositor system.
package markdown

import (
	"github.com/odvcencio/furry-ui/compositor"
	"github.com/odvcencio/furry-ui/theme"
)

// StyledSpan represents a span of text with consistent styling.
type StyledSpan struct {
	Text  string
	Style compositor.Style
}

// StyledLine represents a line composed of styled spans.
type StyledLine struct {
	Spans                  []StyledSpan
	Prefix                 []StyledSpan // Repeated prefix for wrapped lines (e.g., list bullets)
	Indent                 int          // Indentation level for nested structures
	IsCode                 bool         // Inside a code block
	IsCodeHeader           bool         // Header line for a code block
	Language               string       // Language for syntax highlighting
	CodeLineNumberWidth    int          // Width of the code line number gutter
	CodeLineNumberOptional bool         // Line numbers appear on hover for short blocks
	BlankLine              bool         // Empty line for spacing
}

// Plain creates a StyledLine from plain text with default style.
func Plain(text string, style compositor.Style) StyledLine {
	return StyledLine{
		Spans: []StyledSpan{{Text: text, Style: style}},
	}
}

// Blank creates an empty line for spacing.
func Blank() StyledLine {
	return StyledLine{BlankLine: true}
}

// StyleConfig maps markdown elements to compositor styles.
type StyleConfig struct {
	// Headers
	H1 compositor.Style
	H2 compositor.Style
	H3 compositor.Style
	H4 compositor.Style
	H5 compositor.Style
	H6 compositor.Style

	// Inline styles
	Bold          compositor.Style
	Italic        compositor.Style
	BoldItalic    compositor.Style
	Code          compositor.Style
	Strikethrough compositor.Style
	Link          compositor.Style
	LinkURL       compositor.Style

	// Block elements
	Blockquote       compositor.Style
	BlockquoteBorder compositor.Style
	ListBullet       compositor.Style
	ListNumber       compositor.Style
	HorizontalRule   compositor.Style

	// Code blocks
	CodeBlockBorder     compositor.Style
	CodeBlockBG         compositor.Style
	CodeBlockLang       compositor.Style
	CodeBlockLineNumber compositor.Style

	// Tables
	TableHeader compositor.Style
	TableCell   compositor.Style
	TableBorder compositor.Style

	// Base text
	Text compositor.Style
}

// DefaultStyleConfig returns style configuration based on the theme.
func DefaultStyleConfig(t *theme.Theme) *StyleConfig {
	if t == nil {
		t = theme.DefaultTheme()
	}

	// Base text color
	textFG := t.TextPrimary.FG

	return &StyleConfig{
		// Headers - accent color, decreasing emphasis
		H1: compositor.DefaultStyle().WithFG(t.Accent.FG).WithBold(true),
		H2: compositor.DefaultStyle().WithFG(textFG).WithBold(true),
		H3: compositor.DefaultStyle().WithFG(textFG).WithBold(true).WithDim(true),
		H4: compositor.DefaultStyle().WithFG(t.TextSecondary.FG).WithBold(true),
		H5: compositor.DefaultStyle().WithFG(t.TextSecondary.FG),
		H6: compositor.DefaultStyle().WithFG(t.TextMuted.FG),

		// Inline styles - inherit base and add attributes
		Bold:          compositor.DefaultStyle().WithFG(textFG).WithBold(true),
		Italic:        compositor.DefaultStyle().WithFG(textFG).WithItalic(true),
		BoldItalic:    compositor.DefaultStyle().WithFG(textFG).WithBold(true).WithItalic(true),
		Code:          compositor.DefaultStyle().WithFG(t.Tool.FG).WithBG(t.Surface.BG),
		Strikethrough: compositor.DefaultStyle().WithFG(t.TextMuted.FG).WithDim(true),
		Link:          compositor.DefaultStyle().WithFG(t.Info.FG).WithUnderline(true),
		LinkURL:       compositor.DefaultStyle().WithFG(t.TextMuted.FG),

		// Block elements
		Blockquote:       compositor.DefaultStyle().WithFG(t.TextSecondary.FG).WithItalic(true),
		BlockquoteBorder: compositor.DefaultStyle().WithFG(t.Border.FG),
		ListBullet:       compositor.DefaultStyle().WithFG(t.Accent.FG),
		ListNumber:       compositor.DefaultStyle().WithFG(t.Accent.FG),
		HorizontalRule:   compositor.DefaultStyle().WithFG(t.Border.FG),

		// Code blocks
		CodeBlockBorder:     compositor.DefaultStyle().WithFG(t.Accent.FG),
		CodeBlockBG:         compositor.DefaultStyle().WithBG(t.Surface.BG),
		CodeBlockLang:       compositor.DefaultStyle().WithFG(t.TextMuted.FG).WithItalic(true),
		CodeBlockLineNumber: compositor.DefaultStyle().WithFG(t.TextMuted.FG),

		// Tables
		TableHeader: compositor.DefaultStyle().WithFG(textFG).WithBold(true),
		TableCell:   compositor.DefaultStyle().WithFG(textFG),
		TableBorder: compositor.DefaultStyle().WithFG(t.Border.FG),

		// Base
		Text: compositor.DefaultStyle().WithFG(textFG),
	}
}

// WithBaseText returns a copy of the config with base text styles overridden.
func (c *StyleConfig) WithBaseText(base compositor.Style) *StyleConfig {
	if c == nil {
		return c
	}
	next := *c
	next.Text = base
	next.Bold = base.WithBold(true)
	next.Italic = base.WithItalic(true)
	next.BoldItalic = base.WithBold(true).WithItalic(true)
	return &next
}

// MergeStyle combines a base style with inline style attributes.
func MergeStyle(base, inline compositor.Style) compositor.Style {
	result := base
	if inline.Bold {
		result.Bold = true
	}
	if inline.Italic {
		result.Italic = true
	}
	if inline.Underline {
		result.Underline = true
	}
	if inline.Strikethrough {
		result.Strikethrough = true
	}
	if inline.Dim {
		result.Dim = true
	}
	if inline.FG.Mode != compositor.ColorModeDefault && inline.FG.Mode != compositor.ColorModeNone {
		result.FG = inline.FG
	}
	if inline.BG.Mode != compositor.ColorModeDefault && inline.BG.Mode != compositor.ColorModeNone {
		result.BG = inline.BG
	}
	return result
}
