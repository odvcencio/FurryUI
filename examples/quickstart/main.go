package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/odvcencio/furry-ui/backend"
	backendtcell "github.com/odvcencio/furry-ui/backend/tcell"
	"github.com/odvcencio/furry-ui/runtime"
	"github.com/odvcencio/furry-ui/state"
	"github.com/odvcencio/furry-ui/widgets"
)

func main() {
	be, err := backendtcell.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "backend init failed: %v\n", err)
		os.Exit(1)
	}

	app := runtime.NewApp(runtime.AppConfig{
		Backend:  be,
		TickRate: time.Second / 30,
	})

	count := state.NewSignal(0)
	count.SetEqualFunc(state.EqualComparable[int])

	mode := state.NewSignal("manual")
	mode.SetEqualFunc(state.EqualComparable[string])

	root := NewCounterView(count, mode)
	app.SetRoot(root)

	app.Every(400*time.Millisecond, func(now time.Time) runtime.Message {
		if mode != nil && mode.Get() == "auto" {
			count.Update(func(v int) int { return v + 1 })
		}
		return nil
	})

	if err := app.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

type CounterView struct {
	widgets.Component
	count        *state.Signal[int]
	mode         *state.Signal[string]
	countValue   int
	modeValue    string
	spinnerIndex int
	spinner      []rune
}

func NewCounterView(count *state.Signal[int], mode *state.Signal[string]) *CounterView {
	view := &CounterView{
		count:   count,
		mode:    mode,
		spinner: []rune{'|', '/', '-', '\\'},
	}
	view.refresh()
	return view
}

func (c *CounterView) Mount() {
	c.Observe(c.count, c.refresh)
	c.Observe(c.mode, c.refresh)
	c.refresh()
}

func (c *CounterView) Unmount() {
	c.Subs.Clear()
}

func (c *CounterView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (c *CounterView) Layout(bounds runtime.Rect) {
	c.Component.Layout(bounds)
}

func (c *CounterView) Render(ctx runtime.RenderContext) {
	if ctx.Buffer == nil {
		return
	}
	bounds := c.Bounds()
	if bounds.Width == 0 || bounds.Height == 0 {
		return
	}
	ctx.Clear(backend.DefaultStyle())

	frame := c.spinner[c.spinnerIndex%len(c.spinner)]
	lines := []string{
		"[" + string(frame) + "] FurryUI Quickstart",
		"",
		"Count: " + strconv.Itoa(c.countValue),
		"Mode: " + c.modeValue,
		"",
		"Keys: +/- to change, m to toggle auto, q to quit",
	}

	for i, line := range lines {
		if i >= bounds.Height {
			break
		}
		if len(line) > bounds.Width {
			line = line[:bounds.Width]
		}
		ctx.Buffer.SetString(bounds.X, bounds.Y+i, line, backend.DefaultStyle())
	}
}

func (c *CounterView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	switch m := msg.(type) {
	case runtime.KeyMsg:
		switch m.Rune {
		case 'q':
			return runtime.WithCommand(runtime.Quit{})
		case '+', '=':
			c.updateCount(1)
			return runtime.Handled()
		case '-':
			c.updateCount(-1)
			return runtime.Handled()
		case 'm':
			c.toggleMode()
			return runtime.Handled()
		}
	case runtime.TickMsg:
		if len(c.spinner) > 0 {
			c.spinnerIndex = (c.spinnerIndex + 1) % len(c.spinner)
			c.Invalidate()
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (c *CounterView) refresh() {
	if c.count != nil {
		c.countValue = c.count.Get()
	}
	if c.mode != nil {
		c.modeValue = c.mode.Get()
	}
}

func (c *CounterView) updateCount(delta int) {
	if c.count == nil {
		return
	}
	c.count.Update(func(v int) int { return v + delta })
}

func (c *CounterView) toggleMode() {
	if c.mode == nil {
		return
	}
	if c.mode.Get() == "auto" {
		c.mode.Set("manual")
		return
	}
	c.mode.Set("auto")
}
