package markdown

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/odvcencio/furry-ui/compositor"
	"github.com/odvcencio/furry-ui/theme"
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

// Renderer turns markdown into styled lines for the TUI.
type Renderer struct {
	parser      *Parser
	baseConfig  *StyleConfig
	theme       *theme.Theme
	highlighter *Highlighter
}

// NewRenderer creates a renderer using the provided theme.
func NewRenderer(t *theme.Theme) *Renderer {
	if t == nil {
		t = theme.DefaultTheme()
	}
	cfg := DefaultStyleConfig(t)
	return &Renderer{
		parser:      NewParser(),
		baseConfig:  cfg,
		theme:       t,
		highlighter: NewHighlighter(t),
	}
}

// Render parses markdown and returns styled lines for the given source.
func (r *Renderer) Render(source, content string) []StyledLine {
	cfg := r.configForSource(source)
	return r.renderWithConfig(content, cfg)
}

// CodeBlockBackground returns the default code block background style.
func (r *Renderer) CodeBlockBackground() compositor.Style {
	if r == nil || r.baseConfig == nil {
		return compositor.DefaultStyle()
	}
	return r.baseConfig.CodeBlockBG
}

func (r *Renderer) configForSource(source string) *StyleConfig {
	cfg := r.baseConfig
	if cfg == nil {
		cfg = DefaultStyleConfig(r.theme)
	}
	base := baseStyleForSource(source, r.theme)
	return cfg.WithBaseText(base)
}

func baseStyleForSource(source string, t *theme.Theme) compositor.Style {
	if t == nil {
		t = theme.DefaultTheme()
	}
	switch source {
	case "user":
		return t.User
	case "assistant":
		return t.Assistant
	case "system":
		return t.System
	case "tool":
		return t.Tool
	case "thinking":
		return t.Thinking
	default:
		return t.TextPrimary
	}
}

func (r *Renderer) renderWithConfig(content string, cfg *StyleConfig) []StyledLine {
	if cfg == nil {
		return nil
	}
	root := r.parser.ParseString(content)
	state := newRenderState(cfg, []byte(content), r.highlighter)

	for node := root.FirstChild(); node != nil; node = node.NextSibling() {
		r.renderBlock(node, state, false)
	}
	state.flushLine(false, false, "")
	state.trimTrailingBlankLines()
	return state.lines
}

type renderState struct {
	cfg         *StyleConfig
	source      []byte
	baseStyle   compositor.Style
	lines       []StyledLine
	current     []StyledSpan
	prefix      []StyledSpan
	highlighter *Highlighter
}

func newRenderState(cfg *StyleConfig, source []byte, highlighter *Highlighter) *renderState {
	return &renderState{
		cfg:         cfg,
		source:      source,
		baseStyle:   cfg.Text,
		highlighter: highlighter,
	}
}

func (s *renderState) appendSpan(span StyledSpan) {
	if span.Text == "" {
		return
	}
	if len(s.current) > 0 {
		last := &s.current[len(s.current)-1]
		if last.Style.Equal(span.Style) {
			last.Text += span.Text
			return
		}
	}
	s.current = append(s.current, span)
}

func (s *renderState) appendText(text string, style compositor.Style) {
	s.appendSpan(StyledSpan{Text: text, Style: style})
}

func (s *renderState) flushLine(force bool, isCode bool, language string) {
	if len(s.current) == 0 && !force {
		return
	}
	line := StyledLine{
		Spans:    s.current,
		Prefix:   append([]StyledSpan{}, s.prefix...),
		IsCode:   isCode,
		Language: language,
	}
	if len(s.current) == 0 {
		line.BlankLine = true
	}
	s.lines = append(s.lines, line)
	s.current = nil
}

func (s *renderState) addSpacer() {
	if len(s.lines) > 0 && s.lines[len(s.lines)-1].BlankLine {
		return
	}
	s.lines = append(s.lines, StyledLine{BlankLine: true})
}

func (s *renderState) trimTrailingBlankLines() {
	for len(s.lines) > 0 && s.lines[len(s.lines)-1].BlankLine {
		s.lines = s.lines[:len(s.lines)-1]
	}
}

func (s *renderState) withPrefix(extra []StyledSpan, fn func()) {
	prev := s.prefix
	if len(extra) > 0 {
		combined := make([]StyledSpan, 0, len(prev)+len(extra))
		combined = append(combined, prev...)
		combined = append(combined, extra...)
		s.prefix = combined
	}
	fn()
	s.prefix = prev
}

func (s *renderState) withBaseStyle(base compositor.Style, fn func()) {
	prev := s.baseStyle
	s.baseStyle = base
	fn()
	s.baseStyle = prev
}

func (r *Renderer) renderBlock(node ast.Node, state *renderState, tight bool) {
	switch n := node.(type) {
	case *ast.Paragraph:
		r.renderInlineChildren(n, state, state.baseStyle)
		state.flushLine(false, false, "")
		if !tight {
			state.addSpacer()
		}

	case *ast.Heading:
		style := headingStyle(state.cfg, n.Level)
		r.renderInlineChildren(n, state, style)
		state.flushLine(false, false, "")
		state.addSpacer()

	case *ast.Blockquote:
		prefix := []StyledSpan{{Text: "│ ", Style: state.cfg.BlockquoteBorder}}
		blockStyle := MergeStyle(state.baseStyle, state.cfg.Blockquote)
		state.withPrefix(prefix, func() {
			state.withBaseStyle(blockStyle, func() {
				for child := n.FirstChild(); child != nil; child = child.NextSibling() {
					r.renderBlock(child, state, tight)
				}
			})
		})
		state.addSpacer()

	case *ast.List:
		r.renderList(n, state)
		if !tight {
			state.addSpacer()
		}

	case *ast.FencedCodeBlock:
		r.renderCodeBlock(state, n.Text(state.source), string(n.Language(state.source)))
		state.addSpacer()

	case *ast.CodeBlock:
		r.renderCodeBlock(state, n.Text(state.source), "")
		state.addSpacer()

	case *ast.ThematicBreak:
		state.flushLine(false, false, "")
		state.appendText(strings.Repeat("-", 32), state.cfg.HorizontalRule)
		state.flushLine(true, false, "")
		state.addSpacer()

	case *extast.Table:
		r.renderTable(n, state)
		state.addSpacer()

	default:
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			r.renderBlock(child, state, tight)
		}
	}
}

func (r *Renderer) renderList(list *ast.List, state *renderState) {
	start := list.Start
	if start == 0 {
		start = 1
	}
	index := start
	depth := listDepth(list)

	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		prefix := listPrefix(state, list, index, depth)
		state.withPrefix(prefix, func() {
			if li, ok := item.(*ast.ListItem); ok {
				for child := li.FirstChild(); child != nil; child = child.NextSibling() {
					r.renderBlock(child, state, list.IsTight)
				}
			}
		})
		index++
		if !list.IsTight {
			state.addSpacer()
		}
	}
}

func listDepth(list *ast.List) int {
	depth := 0
	for node := list.Parent(); node != nil; node = node.Parent() {
		if _, ok := node.(*ast.List); ok {
			depth++
		}
	}
	return depth
}

func listPrefix(state *renderState, list *ast.List, index, depth int) []StyledSpan {
	indent := strings.Repeat("  ", depth)
	base := state.baseStyle
	bulletStyle := state.cfg.ListBullet
	bullet := theme.Symbols.Bullet
	if list.IsOrdered() {
		bulletStyle = state.cfg.ListNumber
		bullet = fmt.Sprintf("%d.", index)
	}
	var spans []StyledSpan
	if indent != "" {
		spans = append(spans, StyledSpan{Text: indent, Style: base})
	}
	spans = append(spans, StyledSpan{Text: bullet, Style: bulletStyle})
	spans = append(spans, StyledSpan{Text: " ", Style: base})
	return spans
}

func headingStyle(cfg *StyleConfig, level int) compositor.Style {
	switch level {
	case 1:
		return cfg.H1
	case 2:
		return cfg.H2
	case 3:
		return cfg.H3
	case 4:
		return cfg.H4
	case 5:
		return cfg.H5
	default:
		return cfg.H6
	}
}

func (r *Renderer) renderCodeBlock(state *renderState, raw []byte, language string) {
	state.flushLine(false, false, "")
	code := strings.TrimRight(string(raw), "\n")

	label := language
	if label == "" {
		label = "code"
	}
	header := label + "  [copy] [open]"
	basePrefix := append([]StyledSpan{}, state.prefix...)
	headerPrefix := append(basePrefix, StyledSpan{Text: "│ ", Style: state.cfg.CodeBlockBorder})
	state.lines = append(state.lines, StyledLine{
		Spans:        []StyledSpan{{Text: header, Style: state.cfg.CodeBlockLang}},
		Prefix:       headerPrefix,
		IsCode:       true,
		IsCodeHeader: true,
		Language:     language,
	})

	var lines []StyledLine
	if state.highlighter != nil {
		lines = state.highlighter.Highlight(code, language, state.cfg)
	} else {
		lines = fallbackCodeLines(code, language, state.cfg)
	}

	lineCount := len(lines)
	showLineNumbers := lineCount > 0
	optionalNumbers := lineCount <= 10
	numWidth := 0
	numberStyle := state.cfg.CodeBlockLineNumber
	borderStyle := state.cfg.CodeBlockBorder
	if showLineNumbers {
		numWidth = len(strconv.Itoa(lineCount))
		numberStyle = applyCodeBlockBG(numberStyle, state.cfg.CodeBlockBG)
		borderStyle = applyCodeBlockBG(borderStyle, state.cfg.CodeBlockBG)
	}

	for i := range lines {
		linePrefix := append([]StyledSpan{}, basePrefix...)
		if showLineNumbers {
			num := fmt.Sprintf("%*d", numWidth, i+1)
			linePrefix = append(linePrefix,
				StyledSpan{Text: num, Style: numberStyle},
				StyledSpan{Text: " ", Style: numberStyle},
				StyledSpan{Text: "│ ", Style: borderStyle},
			)
		} else {
			linePrefix = append(linePrefix, StyledSpan{Text: "│ ", Style: state.cfg.CodeBlockBorder})
		}
		lines[i].Prefix = append(linePrefix, lines[i].Prefix...)
		if showLineNumbers {
			lines[i].CodeLineNumberWidth = numWidth
			lines[i].CodeLineNumberOptional = optionalNumbers
		}
		state.lines = append(state.lines, lines[i])
	}
}

func fallbackCodeLines(code, language string, cfg *StyleConfig) []StyledLine {
	if code == "" {
		return []StyledLine{{IsCode: true, Language: language, BlankLine: true}}
	}
	parts := strings.Split(code, "\n")
	lines := make([]StyledLine, 0, len(parts))
	style := cfg.Code
	style = applyCodeBlockBG(style, cfg.CodeBlockBG)
	for _, part := range parts {
		line := StyledLine{
			Spans:    []StyledSpan{{Text: part, Style: style}},
			IsCode:   true,
			Language: language,
		}
		if part == "" {
			line.BlankLine = true
		}
		lines = append(lines, line)
	}
	return lines
}

func (r *Renderer) renderTable(table *extast.Table, state *renderState) {
	for row := table.FirstChild(); row != nil; row = row.NextSibling() {
		header, isHeader := row.(*extast.TableHeader)
		var rowNode ast.Node = row
		if isHeader {
			rowNode = header
		}
		var cells []StyledSpan
		for cell := rowNode.FirstChild(); cell != nil; cell = cell.NextSibling() {
			text := collectPlainText(cell, state.source)
			cellStyle := state.cfg.TableCell
			if isHeader {
				cellStyle = state.cfg.TableHeader
			}
			cells = append(cells, StyledSpan{Text: text, Style: cellStyle})
			if cell.NextSibling() != nil {
				cells = append(cells, StyledSpan{Text: " | ", Style: state.cfg.TableBorder})
			}
		}
		state.lines = append(state.lines, StyledLine{
			Spans:  cells,
			Prefix: append([]StyledSpan{}, state.prefix...),
		})
	}
}

func (r *Renderer) renderInlineChildren(node ast.Node, state *renderState, style compositor.Style) {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		r.renderInline(child, state, style)
	}
}

func (r *Renderer) renderInline(node ast.Node, state *renderState, style compositor.Style) {
	switch n := node.(type) {
	case *ast.Text:
		text := string(n.Segment.Value(state.source))
		if text != "" {
			state.appendText(text, style)
		}
		if n.SoftLineBreak() {
			state.appendText(" ", style)
		}
		if n.HardLineBreak() {
			state.flushLine(true, false, "")
		}

	case *ast.String:
		text := string(n.Value)
		if text != "" {
			state.appendText(text, style)
		}

	case *ast.CodeSpan:
		code := collectPlainText(n, state.source)
		state.appendText(code, state.cfg.Code)

	case *ast.Emphasis:
		inline := state.cfg.Italic
		if n.Level >= 2 {
			inline = state.cfg.Bold
		}
		merged := MergeStyle(style, inline)
		r.renderInlineChildren(n, state, merged)

	case *extast.Strikethrough:
		merged := MergeStyle(style, state.cfg.Strikethrough)
		r.renderInlineChildren(n, state, merged)

	case *ast.Link:
		merged := MergeStyle(style, state.cfg.Link)
		r.renderInlineChildren(n, state, merged)
		dest := string(n.Destination)
		label := collectPlainText(n, state.source)
		if dest != "" && dest != label {
			state.appendText(" ("+dest+")", state.cfg.LinkURL)
		}

	case *ast.Image:
		merged := MergeStyle(style, state.cfg.Link)
		r.renderInlineChildren(n, state, merged)
		dest := string(n.Destination)
		if dest != "" {
			state.appendText(" ("+dest+")", state.cfg.LinkURL)
		}

	case *ast.AutoLink:
		url := string(n.URL(state.source))
		state.appendText(url, MergeStyle(style, state.cfg.Link))

	case *extast.TaskCheckBox:
		box := "[ ] "
		if n.IsChecked {
			box = "[x] "
		}
		state.appendText(box, state.cfg.ListBullet)

	default:
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			r.renderInline(child, state, style)
		}
	}
}

func collectPlainText(node ast.Node, source []byte) string {
	var b strings.Builder
	var walk func(ast.Node)
	walk = func(n ast.Node) {
		switch t := n.(type) {
		case *ast.Text:
			b.Write(t.Segment.Value(source))
			if t.SoftLineBreak() {
				b.WriteByte(' ')
			}
			if t.HardLineBreak() {
				b.WriteByte('\n')
			}
		case *ast.String:
			b.Write(t.Value)
		}
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			walk(child)
		}
	}
	walk(node)
	return b.String()
}
