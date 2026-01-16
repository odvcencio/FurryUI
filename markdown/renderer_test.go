package markdown

import (
	"strings"
	"testing"

	"github.com/odvcencio/furry-ui/theme"
)

func TestRenderer_RenderParagraph(t *testing.T) {
	r := NewRenderer(theme.DefaultTheme())
	lines := r.Render("assistant", "Hello **world**")
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}
	got := spansText(lines[0].Spans)
	if got != "Hello world" {
		t.Fatalf("got %q, want %q", got, "Hello world")
	}
}

func TestRenderer_RenderCodeBlock(t *testing.T) {
	r := NewRenderer(theme.DefaultTheme())
	md := "```go\nfmt.Println(\"hi\")\n```\n"
	lines := r.Render("assistant", md)
	if len(lines) == 0 {
		t.Fatal("expected at least one line")
	}

	var codeLines []StyledLine
	for _, line := range lines {
		if line.IsCode {
			codeLines = append(codeLines, line)
		}
	}
	if len(codeLines) == 0 {
		t.Fatal("expected code lines")
	}

	var found bool
	for _, line := range codeLines {
		if strings.Contains(spansText(line.Spans), "fmt.Println") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected highlighted code content")
	}
}

func spansText(spans []StyledSpan) string {
	var b strings.Builder
	for _, span := range spans {
		b.WriteString(span.Text)
	}
	return b.String()
}
