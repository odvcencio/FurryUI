// Demo Generator - Creates asciicast recordings of FluffyUI widgets
//
// This tool generates demo recordings using the simulation backend,
// which doesn't require a real terminal. Perfect for CI/CD pipelines.
//
// Usage:
//
//	go run ./examples/generate-demos --out docs/demos
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/backend/sim"
	"github.com/odvcencio/fluffy-ui/recording"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
	"github.com/odvcencio/fluffy-ui/widgets"
)

var (
	outDir = flag.String("out", "docs/demos", "output directory for recordings")
	width  = flag.Int("width", 80, "recording width")
	height = flag.Int("height", 24, "recording height")
)

func main() {
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	demos := []struct {
		name string
		fn   func() runtime.Widget
	}{
		{"buttons", demoButtons},
		{"counter", demoCounter},
		{"table", demoTable},
		{"progress", demoProgress},
		{"list", demoList},
		{"dialog", demoDialog},
		{"sparkline", demoSparkline},
		{"tabs", demoTabs},
		{"hero", demoHero},
	}

	for _, demo := range demos {
		outPath := filepath.Join(*outDir, demo.name+".cast")
		fmt.Printf("Recording: %s -> %s\n", demo.name, outPath)

		if err := recordDemo(outPath, demo.fn()); err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR: %v\n", err)
			continue
		}
		fmt.Println("  OK")
	}

	fmt.Println("\nDemos recorded successfully!")
	fmt.Println("\nTo view recordings:")
	fmt.Printf("  asciinema play %s/hero.cast\n", *outDir)
	fmt.Println("\nTo convert to GIF (requires agg):")
	fmt.Printf("  agg --theme monokai --last-frame-duration 0.001 %s/hero.cast %s/hero.gif\n", *outDir, *outDir)
}

func recordDemo(path string, root runtime.Widget) error {
	recorder, err := recording.NewAsciicastRecorder(path, recording.AsciicastOptions{
		Title: "FluffyUI Demo",
	})
	if err != nil {
		return err
	}

	frameCount := 0
	maxFrames := 90 // 3 seconds at 30fps

	update := func(app *runtime.App, msg runtime.Message) bool {
		switch msg.(type) {
		case runtime.TickMsg:
			frameCount++
			if frameCount >= maxFrames {
				app.ExecuteCommand(runtime.Quit{})
				return false
			}
			// Forward tick to widgets so animations work
			runtime.DefaultUpdate(app, msg)
			return true
		}
		return runtime.DefaultUpdate(app, msg)
	}

	app := runtime.NewApp(runtime.AppConfig{
		Backend:  sim.New(*width, *height),
		Root:     root,
		Update:   update,
		TickRate: time.Second / 30,
		Recorder: recorder,
	})

	return app.Run(context.Background())
}

// =============================================================================
// Demo Widgets
// =============================================================================

func demoButtons() runtime.Widget {
	return &buttonsDemo{}
}

type buttonsDemo struct {
	widgets.Component
	frame   int
	focused int
}

func (b *buttonsDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (b *buttonsDemo) Layout(bounds runtime.Rect) {
	b.Component.Layout(bounds)
}

func (b *buttonsDemo) Render(ctx runtime.RenderContext) {
	bounds := b.Bounds()
	ctx.Clear(backend.DefaultStyle())

	// Title
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, "FluffyUI Button Variants", backend.DefaultStyle().Bold(true))

	// Button definitions with colors
	buttons := []struct {
		label string
		style backend.Style
	}{
		{"Primary", backend.DefaultStyle().Foreground(backend.ColorBlack).Background(backend.ColorCyan).Bold(true)},
		{"Secondary", backend.DefaultStyle().Foreground(backend.ColorWhite).Background(backend.ColorBlue)},
		{"Danger", backend.DefaultStyle().Foreground(backend.ColorWhite).Background(backend.ColorRed).Bold(true)},
		{"Success", backend.DefaultStyle().Foreground(backend.ColorBlack).Background(backend.ColorGreen).Bold(true)},
		{"Warning", backend.DefaultStyle().Foreground(backend.ColorBlack).Background(backend.ColorYellow)},
	}

	y := bounds.Y + 3
	x := bounds.X + 2
	for i, btn := range buttons {
		style := btn.style
		// Highlight focused button
		if i == b.focused {
			style = style.Reverse(true)
		}
		label := fmt.Sprintf(" [%s] ", btn.label)
		ctx.Buffer.SetString(x, y, label, style)
		x += len(label) + 2
	}

	// Second row: default and disabled
	y += 2
	x = bounds.X + 2

	defaultStyle := backend.DefaultStyle()
	if b.focused == 5 {
		defaultStyle = defaultStyle.Reverse(true)
	}
	ctx.Buffer.SetString(x, y, " [Default] ", defaultStyle)
	x += 13

	disabledStyle := backend.DefaultStyle().Dim(true)
	ctx.Buffer.SetString(x, y, " [Disabled] ", disabledStyle)

	// Instructions
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+7, "Tab cycles focus, Enter activates", backend.DefaultStyle().Dim(true))

	// Focus indicator
	focusText := fmt.Sprintf("Focus: %d/6", b.focused+1)
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+9, focusText, backend.DefaultStyle())

	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func (b *buttonsDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		b.frame++
		// Cycle focus every ~500ms (15 frames at 30fps)
		if b.frame%15 == 0 {
			b.focused = (b.focused + 1) % 6
			b.Invalidate()
		}
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func demoCounter() runtime.Widget {
	count := state.NewSignal(0)
	count.SetEqualFunc(state.EqualComparable[int])

	return &counterDemo{count: count}
}

type counterDemo struct {
	widgets.Component
	count *state.Signal[int]
	frame int
}

func (c *counterDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (c *counterDemo) Layout(bounds runtime.Rect) {
	c.Component.Layout(bounds)
}

func (c *counterDemo) Render(ctx runtime.RenderContext) {
	bounds := c.Bounds()
	ctx.Clear(backend.DefaultStyle())

	// Title
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, "Reactive State Demo", backend.DefaultStyle().Bold(true))

	// Count display
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+3, fmt.Sprintf("Count: %d", c.count.Get()), backend.DefaultStyle())

	// Auto-increment animation
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+5, "Auto-incrementing with signals...", backend.DefaultStyle().Dim(true))

	// Draw border
	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func (c *counterDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		c.frame++
		if c.frame%10 == 0 { // Every ~333ms
			c.count.Update(func(v int) int { return v + 1 })
			c.Invalidate()
		}
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func demoTable() runtime.Widget {
	return &tableDemo{}
}

type tableDemo struct {
	widgets.Component
	frame    int
	selected int
}

func (t *tableDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (t *tableDemo) Layout(bounds runtime.Rect) {
	t.Component.Layout(bounds)
}

func (t *tableDemo) Render(ctx runtime.RenderContext) {
	bounds := t.Bounds()
	ctx.Clear(backend.DefaultStyle())

	// Column definitions
	columns := []struct {
		title string
		width int
	}{
		{"Name", 15},
		{"Type", 12},
		{"Price", 10},
		{"Stock", 8},
	}

	rows := [][]string{
		{"Gummy Bears", "Candy", "$2.99", "150"},
		{"Chocolate Bar", "Candy", "$4.50", "89"},
		{"Sour Straws", "Candy", "$1.99", "234"},
		{"Lollipops", "Candy", "$0.99", "500"},
		{"Jawbreakers", "Candy", "$3.25", "45"},
		{"Energy Drink", "Beverage", "$5.99", "120"},
	}

	// Title
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, "Data Table", backend.DefaultStyle().Bold(true))

	// Header
	y := bounds.Y + 3
	x := bounds.X + 2
	headerStyle := backend.DefaultStyle().Bold(true).Underline(true)
	for _, col := range columns {
		title := col.title
		for len(title) < col.width {
			title += " "
		}
		ctx.Buffer.SetString(x, y, title, headerStyle)
		x += col.width + 1
	}

	// Rows
	for rowIdx, row := range rows {
		y++
		x = bounds.X + 2
		style := backend.DefaultStyle()
		if rowIdx == t.selected {
			style = style.Reverse(true).Bold(true)
		}
		for colIdx, col := range columns {
			cell := ""
			if colIdx < len(row) {
				cell = row[colIdx]
			}
			for len(cell) < col.width {
				cell += " "
			}
			ctx.Buffer.SetString(x, y, cell, style)
			x += col.width + 1
		}
	}

	// Navigation hint
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+bounds.Height-3, "Arrow keys to navigate rows", backend.DefaultStyle().Dim(true))

	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func (t *tableDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		t.frame++
		// Cycle selection every ~400ms (12 frames at 30fps)
		if t.frame%12 == 0 {
			t.selected = (t.selected + 1) % 6
			t.Invalidate()
		}
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func demoProgress() runtime.Widget {
	return &progressDemo{}
}

type progressDemo struct {
	widgets.Component
	frame   int
	values  []float64
	current int
}

func (p *progressDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (p *progressDemo) Layout(bounds runtime.Rect) {
	p.Component.Layout(bounds)
}

func (p *progressDemo) Render(ctx runtime.RenderContext) {
	bounds := p.Bounds()
	ctx.Clear(backend.DefaultStyle())

	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, "Progress & Gauges Demo", backend.DefaultStyle().Bold(true))

	// Multiple progress bars at different levels
	progress1 := float64(p.frame%100) / 100.0
	progress2 := float64((p.frame+33)%100) / 100.0
	progress3 := float64((p.frame+66)%100) / 100.0

	y := bounds.Y + 3
	ctx.Buffer.SetString(bounds.X+2, y, "Download:", backend.DefaultStyle())
	widgets.DrawGauge(ctx.Buffer, bounds.X+14, y, 40, progress1, widgets.GaugeStyle{
		FillChar:  '#',
		EmptyChar: '-',
	})
	ctx.Buffer.SetString(bounds.X+56, y, fmt.Sprintf("%3.0f%%", progress1*100), backend.DefaultStyle())

	y += 2
	ctx.Buffer.SetString(bounds.X+2, y, "Upload:  ", backend.DefaultStyle())
	widgets.DrawGauge(ctx.Buffer, bounds.X+14, y, 40, progress2, widgets.GaugeStyle{
		FillChar:  '=',
		EmptyChar: ' ',
	})
	ctx.Buffer.SetString(bounds.X+56, y, fmt.Sprintf("%3.0f%%", progress2*100), backend.DefaultStyle())

	y += 2
	ctx.Buffer.SetString(bounds.X+2, y, "Process: ", backend.DefaultStyle())
	widgets.DrawGauge(ctx.Buffer, bounds.X+14, y, 40, progress3, widgets.GaugeStyle{
		FillChar:  '*',
		EmptyChar: '.',
	})
	ctx.Buffer.SetString(bounds.X+56, y, fmt.Sprintf("%3.0f%%", progress3*100), backend.DefaultStyle())

	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func (p *progressDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		p.frame++
		p.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func demoList() runtime.Widget {
	return &listDemo{}
}

type listDemo struct {
	widgets.Component
	frame    int
	selected int
}

func (l *listDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (l *listDemo) Layout(bounds runtime.Rect) {
	l.Component.Layout(bounds)
}

func (l *listDemo) Render(ctx runtime.RenderContext) {
	bounds := l.Bounds()
	ctx.Clear(backend.DefaultStyle())

	items := []struct {
		icon  string
		label string
		desc  string
	}{
		{"*", "Documents", "Personal files"},
		{"*", "Downloads", "Recent downloads"},
		{"*", "Music", "Audio files"},
		{"*", "Pictures", "Image gallery"},
		{"*", "Videos", "Video collection"},
		{"*", "Projects", "Code repositories"},
	}

	// Title
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, "Selectable List", backend.DefaultStyle().Bold(true))

	y := bounds.Y + 3
	for i, item := range items {
		style := backend.DefaultStyle()
		prefix := "  "
		if i == l.selected {
			style = style.Reverse(true).Bold(true)
			prefix = "> "
		}
		line := fmt.Sprintf("%s%s %s - %s", prefix, item.icon, item.label, item.desc)
		// Pad to full width
		for len(line) < bounds.Width-4 {
			line += " "
		}
		ctx.Buffer.SetString(bounds.X+2, y+i, line, style)
	}

	// Navigation hint
	ctx.Buffer.SetString(bounds.X+2, bounds.Y+bounds.Height-3, "Arrow keys to navigate, Enter to select", backend.DefaultStyle().Dim(true))

	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func (l *listDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		l.frame++
		// Cycle selection every ~400ms (12 frames at 30fps)
		if l.frame%12 == 0 {
			l.selected = (l.selected + 1) % 6
			l.Invalidate()
		}
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func demoDialog() runtime.Widget {
	dialog := widgets.NewDialog(
		"Confirm Action",
		"Are you sure you want to proceed?\nThis action cannot be undone.",
		widgets.DialogButton{Label: "Cancel"},
		widgets.DialogButton{Label: "Confirm"},
	)
	return dialog
}

func demoSparkline() runtime.Widget {
	data := state.NewSignal([]float64{
		10, 15, 20, 18, 25, 30, 28, 35, 40, 38,
		45, 50, 48, 55, 60, 58, 65, 70, 68, 75,
	})

	return &sparklineDemo{data: data}
}

type sparklineDemo struct {
	widgets.Component
	data  *state.Signal[[]float64]
	frame int
}

func (s *sparklineDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (s *sparklineDemo) Layout(bounds runtime.Rect) {
	s.Component.Layout(bounds)
}

func (s *sparklineDemo) Render(ctx runtime.RenderContext) {
	bounds := s.Bounds()
	ctx.Clear(backend.DefaultStyle())

	ctx.Buffer.SetString(bounds.X+2, bounds.Y+1, "Sparkline Chart Demo", backend.DefaultStyle().Bold(true))

	sparkline := widgets.NewSparkline(s.data)
	sparkline.Layout(runtime.Rect{X: bounds.X + 2, Y: bounds.Y + 3, Width: bounds.Width - 4, Height: 1})
	sparkline.Render(ctx)

	ctx.Buffer.SetString(bounds.X+2, bounds.Y+5, "Live data visualization", backend.DefaultStyle().Dim(true))
	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func (s *sparklineDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		s.frame++
		if s.frame%5 == 0 {
			// Add new data point
			s.data.Update(func(d []float64) []float64 {
				newVal := d[len(d)-1] + float64((s.frame%20)-10)
				if newVal < 0 {
					newVal = 0
				}
				if newVal > 100 {
					newVal = 100
				}
				d = append(d[1:], newVal)
				return d
			})
			s.Invalidate()
		}
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func demoTabs() runtime.Widget {
	tabs := widgets.NewTabs(
		widgets.Tab{Title: "Overview", Content: widgets.NewLabel("Welcome to FluffyUI!")},
		widgets.Tab{Title: "Features", Content: widgets.NewLabel("35+ widgets, reactive state, accessibility")},
		widgets.Tab{Title: "Getting Started", Content: widgets.NewLabel("go get github.com/odvcencio/fluffy-ui")},
	)
	return widgets.NewPanel(tabs).WithBorder(backend.DefaultStyle()).WithTitle("Tab Navigation")
}

func demoHero() runtime.Widget {
	return &heroDemo{}
}

type heroDemo struct {
	widgets.Component
	frame int
}

func (h *heroDemo) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (h *heroDemo) Layout(bounds runtime.Rect) {
	h.Component.Layout(bounds)
}

// Rainbow colors for the rotating border
var rainbowColors = []backend.Color{
	backend.ColorBrightRed,
	backend.ColorBrightYellow,
	backend.ColorBrightGreen,
	backend.ColorBrightCyan,
	backend.ColorBrightBlue,
	backend.ColorBrightMagenta,
}

// Border characters - stars, sparkles, diamonds
var borderChars = []rune{'★', '✦', '◆', '✧', '❖', '✶', '◇', '✴', '❋', '✸'}

func (h *heroDemo) Render(ctx runtime.RenderContext) {
	bounds := h.Bounds()
	ctx.Clear(backend.DefaultStyle())

	// Draw rotating rainbow border
	h.drawRainbowBorder(ctx, bounds)

	// ASCII art title with color
	title := []string{
		" _____ _       __  __       _   _ ___ ",
		"|  ___| |_   _ / _|/ _|_   _| | | |_ _|",
		"| |_  | | | | | |_| |_| | | | | | || | ",
		"|  _| | | |_| |  _|  _| |_| | |_| || | ",
		"|_|   |_|\\__,_|_| |_|  \\__, |\\___/|___|",
		"                      |___/           ",
	}

	startY := bounds.Y + 3
	titleColor := rainbowColors[(h.frame/8)%len(rainbowColors)]
	for i, line := range title {
		x := (bounds.Width - len(line)) / 2
		style := backend.DefaultStyle().Bold(true).Foreground(titleColor)
		ctx.Buffer.SetString(x, startY+i, line, style)
	}

	// Subtitle
	subtitle := "A batteries-included TUI framework for Go"
	x := (bounds.Width - len(subtitle)) / 2
	ctx.Buffer.SetString(x, startY+7, subtitle, backend.DefaultStyle().Dim(true))

	// Features (animated) with colorful bullets
	features := []string{
		"35+ Ready-to-Use Widgets",
		"Reactive State Management",
		"Accessibility Built-In",
		"Deterministic Testing",
	}

	featureY := startY + 10
	// Features appear quickly one by one and stay visible
	// At 30fps: frame 0-9 = 0, 10-19 = 1, 20-29 = 2, 30-39 = 3, 40+ = all 4
	visibleFeatures := h.frame / 10
	if visibleFeatures > len(features) {
		visibleFeatures = len(features)
	}

	for i := 0; i < visibleFeatures; i++ {
		fx := (bounds.Width - len(features[i]) - 4) / 2
		bulletColor := rainbowColors[(i+h.frame/10)%len(rainbowColors)]
		bulletStyle := backend.DefaultStyle().Foreground(bulletColor).Bold(true)
		textStyle := backend.DefaultStyle()
		ctx.Buffer.SetString(fx, featureY+i, "★ ", bulletStyle)
		ctx.Buffer.SetString(fx+2, featureY+i, features[i], textStyle)
	}

	// Install command with pulsing highlight
	installY := bounds.Y + bounds.Height - 3
	install := " go get github.com/odvcencio/fluffy-ui "
	ix := (bounds.Width - len(install)) / 2
	installColor := rainbowColors[(h.frame/5)%len(rainbowColors)]
	ctx.Buffer.SetString(ix, installY, install, backend.DefaultStyle().Background(installColor).Foreground(backend.ColorBlack).Bold(true))
}

func (h *heroDemo) drawRainbowBorder(ctx runtime.RenderContext, bounds runtime.Rect) {
	width := bounds.Width
	height := bounds.Height
	if width <= 0 || height <= 0 {
		return
	}
	patternLen := len(borderChars) * len(rainbowColors)
	if patternLen == 0 {
		return
	}

	draw := func(x, y, offset int) {
		pos := (offset + h.frame) % patternLen
		char := borderChars[pos%len(borderChars)]
		color := rainbowColors[(pos/2)%len(rainbowColors)]
		style := backend.DefaultStyle().Foreground(color).Bold(true)
		ctx.Buffer.Set(x, y, char, style)
	}

	// Top edge (left to right)
	for i := 0; i < width; i++ {
		draw(bounds.X+i, bounds.Y, i)
	}

	sideLen := height - 2
	if sideLen < 0 {
		sideLen = 0
	}

	// Right edge (top to bottom)
	for i := 0; i < sideLen; i++ {
		draw(bounds.X+width-1, bounds.Y+1+i, width+i)
	}

	// Bottom edge (right to left)
	if height > 1 {
		base := width + sideLen
		for i := 0; i < width; i++ {
			x := bounds.X + width - 1 - i
			draw(x, bounds.Y+height-1, base+i)
		}
	}

	// Left edge (bottom to top)
	if width > 1 {
		base := width + sideLen + width
		for i := 0; i < sideLen; i++ {
			y := bounds.Y + height - 2 - i
			draw(bounds.X, y, base+i)
		}
	}
}

func (h *heroDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		h.frame++
		h.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}
