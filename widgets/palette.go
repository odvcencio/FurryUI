package widgets

import (
	"strings"

	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/runtime"
	"github.com/odvcencio/furry-ui/terminal"
)

// PaletteItem represents a single item in the palette.
type PaletteItem struct {
	ID          string // Unique identifier
	Category    string // Optional category for grouping (e.g., "Recent", "Files", "Actions")
	Label       string // Display text
	Description string // Optional secondary text
	Shortcut    string // Optional keyboard shortcut hint
	Data        any    // Custom data for the action
}

// PaletteWidget provides a fuzzy-filtering command palette overlay.
type PaletteWidget struct {
	FocusableBase

	query    string
	items    []PaletteItem
	filtered []PaletteItem
	selected int

	// Callbacks
	onSelect func(item PaletteItem)
	filterFn func(item PaletteItem, query string) bool

	// Configuration
	title       string
	placeholder string
	maxVisible  int

	// Styles
	bgStyle       backend.Style
	borderStyle   backend.Style
	titleStyle    backend.Style
	queryStyle    backend.Style
	itemStyle     backend.Style
	selectedStyle backend.Style
	categoryStyle backend.Style
	descStyle     backend.Style
	shortcutStyle backend.Style
}

// NewPaletteWidget creates a new palette widget.
func NewPaletteWidget(title string) *PaletteWidget {
	p := &PaletteWidget{
		title:         title,
		placeholder:   "> ",
		maxVisible:    10,
		bgStyle:       backend.DefaultStyle(),
		borderStyle:   backend.DefaultStyle(),
		titleStyle:    backend.DefaultStyle().Bold(true),
		queryStyle:    backend.DefaultStyle().Bold(true),
		itemStyle:     backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		categoryStyle: backend.DefaultStyle().Foreground(backend.ColorBlue),
		descStyle:     backend.DefaultStyle().Foreground(backend.ColorDefault),
		shortcutStyle: backend.DefaultStyle().Foreground(backend.ColorDefault),
	}
	p.filterFn = p.defaultFilter
	return p
}

// SetItems sets the palette items.
func (p *PaletteWidget) SetItems(items []PaletteItem) {
	p.items = items
	p.updateFiltered()
}

// SetOnSelect sets the callback for item selection.
func (p *PaletteWidget) SetOnSelect(fn func(item PaletteItem)) {
	p.onSelect = fn
}

// SetFilterFn sets a custom filter function.
func (p *PaletteWidget) SetFilterFn(fn func(item PaletteItem, query string) bool) {
	p.filterFn = fn
}

// SetPlaceholder sets the query placeholder text.
func (p *PaletteWidget) SetPlaceholder(placeholder string) {
	p.placeholder = placeholder
}

// SetMaxVisible sets the maximum visible items.
func (p *PaletteWidget) SetMaxVisible(max int) {
	p.maxVisible = max
}

// SetStyles configures the palette appearance.
func (p *PaletteWidget) SetStyles(bg, border, title, query, item, selected, category backend.Style) {
	p.bgStyle = bg
	p.borderStyle = border
	p.titleStyle = title
	p.queryStyle = query
	p.itemStyle = item
	p.selectedStyle = selected
	p.categoryStyle = category
}

// Query returns the current query string.
func (p *PaletteWidget) Query() string {
	return p.query
}

// SelectedItem returns the currently selected item, or nil if none.
func (p *PaletteWidget) SelectedItem() *PaletteItem {
	if p.selected >= 0 && p.selected < len(p.filtered) {
		return &p.filtered[p.selected]
	}
	return nil
}

// updateFiltered refilters items based on current query.
func (p *PaletteWidget) updateFiltered() {
	if p.query == "" {
		p.filtered = p.items
	} else {
		p.filtered = nil
		for _, item := range p.items {
			if p.filterFn(item, p.query) {
				p.filtered = append(p.filtered, item)
			}
		}
	}

	// Clamp selection
	if p.selected >= len(p.filtered) {
		p.selected = len(p.filtered) - 1
	}
	if p.selected < 0 && len(p.filtered) > 0 {
		p.selected = 0
	}
}

// defaultFilter performs case-insensitive substring matching.
func (p *PaletteWidget) defaultFilter(item PaletteItem, query string) bool {
	queryLower := strings.ToLower(query)
	return strings.Contains(strings.ToLower(item.Label), queryLower) ||
		strings.Contains(strings.ToLower(item.Description), queryLower) ||
		strings.Contains(strings.ToLower(item.Category), queryLower)
}

// Measure returns the preferred size.
func (p *PaletteWidget) Measure(constraints runtime.Constraints) runtime.Size {
	width := 60
	if constraints.MaxWidth < width {
		width = constraints.MaxWidth
	}

	// Height: title(1) + query(1) + separator(1) + items(up to maxVisible) + border(2)
	itemCount := len(p.filtered)
	if itemCount > p.maxVisible {
		itemCount = p.maxVisible
	}
	height := 5 + itemCount
	if constraints.MaxHeight < height {
		height = constraints.MaxHeight
	}

	return runtime.Size{Width: width, Height: height}
}

// Layout positions the widget (centered overlay).
func (p *PaletteWidget) Layout(bounds runtime.Rect) {
	size := p.Measure(runtime.Constraints{
		MaxWidth:  bounds.Width,
		MaxHeight: bounds.Height,
	})

	x := bounds.X + (bounds.Width-size.Width)/2
	y := bounds.Y + (bounds.Height-size.Height)/2

	newBounds := runtime.Rect{
		X:      x,
		Y:      y,
		Width:  size.Width,
		Height: size.Height,
	}
	p.Base.Layout(newBounds)
}

// Render draws the palette.
func (p *PaletteWidget) Render(ctx runtime.RenderContext) {
	b := p.bounds
	if b.Width < 20 || b.Height < 5 {
		return
	}

	// Draw background
	ctx.Buffer.Fill(b, ' ', p.bgStyle)

	// Draw border
	p.drawBorder(ctx.Buffer, b)

	// Draw title
	titleX := b.X + (b.Width-len(p.title))/2
	ctx.Buffer.SetString(titleX, b.Y, " "+p.title+" ", p.titleStyle)

	y := b.Y + 1

	// Draw query input
	query := p.placeholder + p.query
	if len(query) > b.Width-4 {
		query = query[:b.Width-4]
	}
	ctx.Buffer.SetString(b.X+2, y, query, p.queryStyle)

	// Draw cursor
	cursorX := b.X + 2 + len(query)
	if cursorX < b.X+b.Width-2 && p.focused {
		ctx.Buffer.Set(cursorX, y, '█', p.queryStyle)
	}
	y++

	// Draw separator
	for x := b.X + 1; x < b.X+b.Width-1; x++ {
		ctx.Buffer.Set(x, y, '─', p.borderStyle)
	}
	y++

	// Draw items
	maxItems := b.Height - 4
	if maxItems > len(p.filtered) {
		maxItems = len(p.filtered)
	}

	lastCategory := ""
	for i := 0; i < maxItems && y < b.Y+b.Height-1; i++ {
		item := p.filtered[i]

		// Draw category header if changed
		if item.Category != "" && item.Category != lastCategory {
			lastCategory = item.Category
			ctx.Buffer.SetString(b.X+2, y, item.Category, p.categoryStyle)
			y++
			if y >= b.Y+b.Height-1 {
				break
			}
		}

		// Draw item
		style := p.itemStyle
		if i == p.selected {
			style = p.selectedStyle
			// Fill entire line for selected
			for x := b.X + 1; x < b.X+b.Width-1; x++ {
				ctx.Buffer.Set(x, y, ' ', style)
			}
		}

		// Label (left-aligned)
		label := item.Label
		maxLabel := b.Width - 6
		if item.Shortcut != "" {
			maxLabel -= len(item.Shortcut) + 2
		}
		if len(label) > maxLabel {
			label = label[:maxLabel-3] + "..."
		}
		ctx.Buffer.SetString(b.X+3, y, label, style)

		// Shortcut (right-aligned)
		if item.Shortcut != "" {
			shortcutX := b.X + b.Width - 2 - len(item.Shortcut)
			shortcutStyle := p.shortcutStyle
			if i == p.selected {
				shortcutStyle = style
			}
			ctx.Buffer.SetString(shortcutX, y, item.Shortcut, shortcutStyle)
		}

		y++
	}

	// Draw item count if more items than visible
	if len(p.filtered) > maxItems {
		countStr := intToStr(len(p.filtered)) + " results"
		ctx.Buffer.SetString(b.X+b.Width-2-len(countStr), b.Y+b.Height-1, countStr, p.borderStyle)
	}
}

// drawBorder draws the palette border.
func (p *PaletteWidget) drawBorder(buf *runtime.Buffer, b runtime.Rect) {
	// Corners
	buf.Set(b.X, b.Y, '╭', p.borderStyle)
	buf.Set(b.X+b.Width-1, b.Y, '╮', p.borderStyle)
	buf.Set(b.X, b.Y+b.Height-1, '╰', p.borderStyle)
	buf.Set(b.X+b.Width-1, b.Y+b.Height-1, '╯', p.borderStyle)

	// Horizontal edges
	for x := b.X + 1; x < b.X+b.Width-1; x++ {
		buf.Set(x, b.Y, '─', p.borderStyle)
		buf.Set(x, b.Y+b.Height-1, '─', p.borderStyle)
	}

	// Vertical edges
	for y := b.Y + 1; y < b.Y+b.Height-1; y++ {
		buf.Set(b.X, y, '│', p.borderStyle)
		buf.Set(b.X+b.Width-1, y, '│', p.borderStyle)
	}
}

// HandleMessage processes keyboard input.
func (p *PaletteWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyEscape:
		return runtime.WithCommand(runtime.PopOverlay{})

	case terminal.KeyEnter:
		if item := p.SelectedItem(); item != nil {
			if p.onSelect != nil {
				p.onSelect(*item)
			}
			return runtime.WithCommands(
				runtime.PaletteSelected{ID: item.ID, Data: item.Data},
				runtime.PopOverlay{},
			)
		}
		return runtime.Handled()

	case terminal.KeyUp:
		if p.selected > 0 {
			p.selected--
		}
		return runtime.Handled()

	case terminal.KeyDown:
		if p.selected < len(p.filtered)-1 {
			p.selected++
		}
		return runtime.Handled()

	case terminal.KeyBackspace:
		if len(p.query) > 0 {
			p.query = p.query[:len(p.query)-1]
			p.updateFiltered()
			return runtime.Handled()
		}
		// Empty query, close palette
		return runtime.WithCommand(runtime.PopOverlay{})

	case terminal.KeyRune:
		p.query += string(key.Rune)
		p.updateFiltered()
		return runtime.Handled()
	}

	return runtime.Unhandled()
}
