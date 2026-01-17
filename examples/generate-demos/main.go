// Demo Generator - Creates asciicast recordings of FluffyUI widgets
//
// This tool generates demo recordings using the simulation backend,
// which doesn't require a real terminal. Perfect for CI/CD pipelines.
//
// Usage:
//   go run ./examples/generate-demos --out docs/demos
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
	fmt.Printf("  agg --theme monokai %s/hero.cast %s/hero.gif\n", *outDir, *outDir)
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
	grid := widgets.NewGrid(4, 3)
	grid.Gap = 1

	grid.Add(widgets.NewLabel("FluffyUI Button Variants").WithStyle(backend.DefaultStyle().Bold(true)), 0, 0, 1, 3)

	grid.Add(widgets.NewButton("Primary", widgets.WithVariant(widgets.VariantPrimary)), 1, 0, 1, 1)
	grid.Add(widgets.NewButton("Secondary", widgets.WithVariant(widgets.VariantSecondary)), 1, 1, 1, 1)
	grid.Add(widgets.NewButton("Danger", widgets.WithVariant(widgets.VariantDanger)), 1, 2, 1, 1)

	grid.Add(widgets.NewButton("Default"), 2, 0, 1, 1)
	disabledState := state.NewSignal(true)
	grid.Add(widgets.NewButton("Disabled", widgets.WithDisabled(disabledState)), 2, 1, 1, 1)

	grid.Add(widgets.NewLabel("Press Tab to navigate, Enter to activate").WithStyle(backend.DefaultStyle().Dim(true)), 3, 0, 1, 3)

	return widgets.NewPanel(grid).WithBorder(backend.DefaultStyle()).WithTitle("Buttons Demo")
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
	table := widgets.NewTable(
		widgets.TableColumn{Title: "Name", Width: 15},
		widgets.TableColumn{Title: "Type", Width: 12},
		widgets.TableColumn{Title: "Price", Width: 10},
		widgets.TableColumn{Title: "Stock", Width: 8},
	)
	table.SetRows([][]string{
		{"Gummy Bears", "Candy", "$2.99", "150"},
		{"Chocolate Bar", "Candy", "$4.50", "89"},
		{"Sour Straws", "Candy", "$1.99", "234"},
		{"Lollipops", "Candy", "$0.99", "500"},
		{"Jawbreakers", "Candy", "$3.25", "45"},
		{"Energy Drink", "Beverage", "$5.99", "120"},
	})

	return widgets.NewPanel(table).WithBorder(backend.DefaultStyle()).WithTitle("Data Table")
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
	items := []string{
		"First Item",
		"Second Item",
		"Third Item",
		"Fourth Item",
		"Fifth Item",
	}

	adapter := widgets.NewSliceAdapter(items, func(item string, index int, selected bool, ctx runtime.RenderContext) {
		style := backend.DefaultStyle()
		if selected {
			style = style.Reverse(true)
		}
		prefix := "  "
		if selected {
			prefix = "> "
		}
		text := prefix + item
		for len(text) < ctx.Bounds.Width {
			text += " "
		}
		ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, text, style)
	})

	list := widgets.NewList(adapter)
	return widgets.NewPanel(list).WithBorder(backend.DefaultStyle()).WithTitle("Selectable List")
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

func (h *heroDemo) Render(ctx runtime.RenderContext) {
	bounds := h.Bounds()
	ctx.Clear(backend.DefaultStyle())

	// ASCII art title
	title := []string{
		" _____ _       __  __       _   _ ___ ",
		"|  ___| |_   _ / _|/ _|_   _| | | |_ _|",
		"| |_  | | | | | |_| |_| | | | | | || | ",
		"|  _| | | |_| |  _|  _| |_| | |_| || | ",
		"|_|   |_|\\__,_|_| |_|  \\__, |\\___/|___|",
		"                      |___/           ",
	}

	startY := bounds.Y + 2
	for i, line := range title {
		x := (bounds.Width - len(line)) / 2
		ctx.Buffer.SetString(x, startY+i, line, backend.DefaultStyle().Bold(true))
	}

	// Subtitle
	subtitle := "A batteries-included TUI framework for Go"
	x := (bounds.Width - len(subtitle)) / 2
	ctx.Buffer.SetString(x, startY+7, subtitle, backend.DefaultStyle().Dim(true))

	// Features (animated)
	features := []string{
		"35+ Ready-to-Use Widgets",
		"Reactive State Management",
		"Accessibility Built-In",
		"Deterministic Testing",
	}

	featureY := startY + 10
	visibleFeatures := (h.frame / 20) % (len(features) + 1)
	if visibleFeatures > len(features) {
		visibleFeatures = len(features)
	}

	for i := 0; i < visibleFeatures; i++ {
		fx := (bounds.Width - len(features[i]) - 4) / 2
		ctx.Buffer.SetString(fx, featureY+i, "[*] "+features[i], backend.DefaultStyle())
	}

	// Install command
	installY := bounds.Y + bounds.Height - 3
	install := "go get github.com/odvcencio/fluffy-ui"
	ix := (bounds.Width - len(install)) / 2
	ctx.Buffer.SetString(ix, installY, install, backend.DefaultStyle().Reverse(true))

	ctx.Buffer.DrawBox(bounds, backend.DefaultStyle())
}

func (h *heroDemo) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if _, ok := msg.(runtime.TickMsg); ok {
		h.frame++
		h.Invalidate()
		return runtime.Handled()
	}
	return runtime.Unhandled()
}
