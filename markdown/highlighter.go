package markdown

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/odvcencio/furry-ui/compositor"
	"github.com/odvcencio/furry-ui/theme"
)

// Highlighter applies syntax highlighting to fenced code blocks.
type Highlighter struct {
	palette codePalette
}

type codePalette struct {
	Default     compositor.Style
	Keyword     compositor.Style
	TypeName    compositor.Style
	Function    compositor.Style
	String      compositor.Style
	Number      compositor.Style
	Comment     compositor.Style
	Operator    compositor.Style
	Punctuation compositor.Style
	Builtin     compositor.Style
	Variable    compositor.Style
	Attribute   compositor.Style
	Tag         compositor.Style
	Error       compositor.Style
}

// NewHighlighter returns a theme-aware highlighter.
func NewHighlighter(t *theme.Theme) *Highlighter {
	if t == nil {
		t = theme.DefaultTheme()
	}
	return &Highlighter{palette: newCodePalette(t)}
}

func newCodePalette(t *theme.Theme) codePalette {
	return codePalette{
		Default:     t.TextPrimary,
		Keyword:     t.Accent.WithBold(true),
		TypeName:    t.Info,
		Function:    t.AccentDim,
		String:      t.Success,
		Number:      t.Warning,
		Comment:     t.TextMuted.WithItalic(true),
		Operator:    t.TextSecondary,
		Punctuation: t.TextMuted,
		Builtin:     t.Info,
		Variable:    t.TextPrimary,
		Attribute:   t.AccentDim,
		Tag:         t.Accent,
		Error:       t.Error.WithBold(true),
	}
}

// Highlight tokenizes and styles code into markdown lines.
func (h *Highlighter) Highlight(code, language string, cfg *StyleConfig) []StyledLine {
	if cfg == nil {
		return fallbackCodeLines(code, language, DefaultStyleConfig(nil))
	}
	if code == "" {
		return []StyledLine{{IsCode: true, Language: language, BlankLine: true}}
	}

	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	iter, err := lexer.Tokenise(nil, code)
	if err != nil {
		return fallbackCodeLines(code, language, cfg)
	}

	var lines []StyledLine
	current := StyledLine{IsCode: true, Language: language}
	flush := func(force bool) {
		if len(current.Spans) == 0 && !force {
			return
		}
		if len(current.Spans) == 0 {
			current.BlankLine = true
		}
		lines = append(lines, current)
		current = StyledLine{IsCode: true, Language: language}
	}

	for token := iter(); token != chroma.EOF; token = iter() {
		if token.Value == "" {
			continue
		}
		style := h.styleForToken(token.Type)
		style = applyCodeBlockBG(style, cfg.CodeBlockBG)
		parts := strings.Split(token.Value, "\n")
		for i, part := range parts {
			if part != "" {
				appendStyledSpan(&current.Spans, StyledSpan{Text: part, Style: style})
			}
			if i < len(parts)-1 {
				flush(true)
			}
		}
	}
	flush(false)

	if len(lines) == 0 {
		lines = append(lines, StyledLine{IsCode: true, Language: language, BlankLine: true})
	}
	return lines
}

func (h *Highlighter) styleForToken(ttype chroma.TokenType) compositor.Style {
	if ttype == chroma.Error {
		return h.palette.Error
	}
	switch {
	case ttype.InCategory(chroma.Comment):
		return h.palette.Comment
	case ttype.InCategory(chroma.Keyword):
		return h.palette.Keyword
	case ttype.InCategory(chroma.LiteralString):
		return h.palette.String
	case ttype.InCategory(chroma.LiteralNumber):
		return h.palette.Number
	case ttype.InCategory(chroma.Operator):
		return h.palette.Operator
	case ttype.InCategory(chroma.Punctuation):
		return h.palette.Punctuation
	case ttype.InCategory(chroma.Name):
		switch ttype {
		case chroma.NameFunction, chroma.NameFunctionMagic:
			return h.palette.Function
		case chroma.NameClass, chroma.NameNamespace:
			return h.palette.TypeName
		case chroma.NameBuiltin, chroma.NameBuiltinPseudo:
			return h.palette.Builtin
		case chroma.NameVariable, chroma.NameVariableClass, chroma.NameVariableGlobal, chroma.NameVariableInstance, chroma.NameVariableMagic:
			return h.palette.Variable
		case chroma.NameTag:
			return h.palette.Tag
		case chroma.NameAttribute:
			return h.palette.Attribute
		case chroma.NameConstant:
			return h.palette.Number
		}
	}
	return h.palette.Default
}

func appendStyledSpan(spans *[]StyledSpan, span StyledSpan) {
	if span.Text == "" {
		return
	}
	if len(*spans) > 0 {
		last := &(*spans)[len(*spans)-1]
		if last.Style.Equal(span.Style) {
			last.Text += span.Text
			return
		}
	}
	*spans = append(*spans, span)
}

func applyCodeBlockBG(style, bg compositor.Style) compositor.Style {
	if bg.BG.Mode != compositor.ColorModeDefault && bg.BG.Mode != compositor.ColorModeNone {
		style = style.WithBG(bg.BG)
	}
	return style
}
