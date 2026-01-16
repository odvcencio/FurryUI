package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Parser wraps goldmark for markdown parsing.
type Parser struct {
	md goldmark.Markdown
}

// NewParser creates a new markdown parser with common extensions enabled.
func NewParser() *Parser {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // GitHub Flavored Markdown (tables, strikethrough, autolinks, task lists)
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	return &Parser{md: md}
}

// Parse parses markdown source and returns the AST root.
func (p *Parser) Parse(source []byte) ast.Node {
	reader := text.NewReader(source)
	return p.md.Parser().Parse(reader)
}

// ParseString is a convenience method for parsing string input.
func (p *Parser) ParseString(source string) ast.Node {
	return p.Parse([]byte(source))
}

// WalkFunc is called for each node during tree traversal.
type WalkFunc func(node ast.Node, entering bool) (ast.WalkStatus, error)

// Walk traverses the AST tree calling fn for each node.
func Walk(node ast.Node, fn WalkFunc) error {
	return ast.Walk(node, ast.Walker(fn))
}
