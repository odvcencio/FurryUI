package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/clipboard"
	"github.com/odvcencio/fluffy-ui/state"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// UpdateFunc handles a message and returns true if a render is needed.
type UpdateFunc func(app *App, msg Message) bool

// CommandHandler handles commands emitted by widgets.
// Return true if the command requires a render.
type CommandHandler func(cmd Command) bool

// AppConfig configures a runtime App.
type AppConfig struct {
	Backend           backend.Backend
	Root              Widget
	Update            UpdateFunc
	CommandHandler    CommandHandler
	MessageBuffer     int
	TickRate          time.Duration
	StateQueue        *state.Queue
	FlushPolicy       QueueFlushPolicy
	KeyHandler        KeyHandler
	Announcer         accessibility.Announcer
	Clipboard         clipboard.Clipboard
	FocusStyle        *accessibility.FocusStyle
	Recorder          Recorder
	RenderObserver    RenderObserver
	FocusRegistration FocusRegistrationMode
}

// App runs a widget tree against a terminal backend.
type App struct {
	backend           backend.Backend
	screen            *Screen
	root              Widget
	update            UpdateFunc
	commandHandler    CommandHandler
	keyHandler        KeyHandler
	messages          chan Message
	tickRate          time.Duration
	stateQueue        *state.Queue
	queueScheduler    *QueueScheduler
	flushPolicy       QueueFlushPolicy
	invalidator       *Invalidator
	announcer         accessibility.Announcer
	clipboard         clipboard.Clipboard
	focusStyle        *accessibility.FocusStyle
	recorder          Recorder
	renderObserver    RenderObserver
	focusRegistration FocusRegistrationMode
	taskCtx           context.Context
	taskCancel        context.CancelFunc
	pendingMu         sync.Mutex
	pendingEffects    []Effect

	running     bool
	dirty       bool
	renderMu    sync.Mutex
	renderFrame int64
}

// NewApp creates a new App from config.
func NewApp(cfg AppConfig) *App {
	bufferSize := cfg.MessageBuffer
	if bufferSize <= 0 {
		bufferSize = 128
	}
	queue := cfg.StateQueue
	if queue == nil {
		queue = state.NewQueue()
	}
	policy := cfg.FlushPolicy
	app := &App{
		backend:           cfg.Backend,
		root:              cfg.Root,
		update:            cfg.Update,
		commandHandler:    cfg.CommandHandler,
		keyHandler:        cfg.KeyHandler,
		messages:          make(chan Message, bufferSize),
		tickRate:          cfg.TickRate,
		stateQueue:        queue,
		flushPolicy:       policy,
		announcer:         cfg.Announcer,
		clipboard:         cfg.Clipboard,
		focusStyle:        cfg.FocusStyle,
		recorder:          cfg.Recorder,
		renderObserver:    cfg.RenderObserver,
		focusRegistration: cfg.FocusRegistration,
	}
	if app.flushPolicy == 0 {
		app.flushPolicy = FlushOnMessageAndTick
	}
	app.queueScheduler = NewQueueScheduler(queue, app.tryPost)
	app.invalidator = NewInvalidator(app.tryPost)
	return app
}

// Screen returns the active screen, if initialized.
func (a *App) Screen() *Screen {
	return a.screen
}

// StateQueue returns the app's state queue.
func (a *App) StateQueue() *state.Queue {
	if a == nil {
		return nil
	}
	return a.stateQueue
}

// StateScheduler returns a scheduler that wakes the app to flush.
func (a *App) StateScheduler() state.Scheduler {
	if a == nil || a.queueScheduler == nil {
		return nil
	}
	return a.queueScheduler
}

// InvalidateScheduler returns a scheduler that invalidates the render pass.
func (a *App) InvalidateScheduler() state.Scheduler {
	if a == nil || a.invalidator == nil {
		return nil
	}
	return a.invalidator
}

// Invalidate requests a render pass.
func (a *App) Invalidate() {
	if a == nil || a.invalidator == nil {
		return
	}
	a.invalidator.Invalidate()
}

// PostQueueFlush requests a state queue flush.
func (a *App) PostQueueFlush() {
	a.Post(QueueFlushMsg{})
}

// Spawn starts an effect using the app task context.
// If Run has not started, the effect is queued until start.
func (a *App) Spawn(effect Effect) {
	if a == nil || effect.Run == nil {
		return
	}
	if a.taskCtx == nil {
		a.pendingMu.Lock()
		a.pendingEffects = append(a.pendingEffects, effect)
		a.pendingMu.Unlock()
		return
	}
	a.runEffect(effect)
}

// After schedules a delayed message using the app task context.
func (a *App) After(delay time.Duration, msg Message) {
	a.Spawn(After(delay, msg))
}

// Every schedules a recurring message using the app task context.
func (a *App) Every(interval time.Duration, fn func(time.Time) Message) {
	a.Spawn(Every(interval, fn))
}

// SetRoot swaps the root widget.
func (a *App) SetRoot(root Widget) {
	a.root = root
	if a.screen != nil {
		a.screen.SetRoot(root)
		a.dirty = true
	}
}

// Post sends a message to the event loop.
func (a *App) Post(msg Message) {
	_ = a.tryPost(msg)
}

// TryPost sends a message to the event loop without blocking.
func (a *App) TryPost(msg Message) bool {
	return a.tryPost(msg)
}

func (a *App) tryPost(msg Message) bool {
	if a == nil || a.messages == nil {
		return false
	}
	select {
	case a.messages <- msg:
		return true
	default:
		return false
	}
}

// Run starts the event loop until quit or context cancellation.
func (a *App) Run(ctx context.Context) error {
	if a.backend == nil {
		return errors.New("backend is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	taskCtx, taskCancel := context.WithCancel(ctx)
	a.taskCtx = taskCtx
	a.taskCancel = taskCancel
	defer func() {
		taskCancel()
		a.taskCtx = nil
		a.taskCancel = nil
	}()
	if err := a.backend.Init(); err != nil {
		return fmt.Errorf("init backend: %w", err)
	}
	defer a.backend.Fini()

	a.backend.HideCursor()
	w, h := a.backend.Size()
	a.screen = NewScreen(w, h)
	a.screen.SetServices(a.Services())
	a.screen.SetAutoRegisterFocus(a.focusRegistration == FocusRegistrationAuto)
	if a.root != nil {
		a.screen.SetRoot(a.root)
	}
	if a.recorder != nil {
		if err := a.recorder.Start(w, h, time.Now()); err != nil {
			return fmt.Errorf("start recorder: %w", err)
		}
		defer func() {
			_ = a.recorder.Close()
		}()
	}

	if a.update == nil {
		a.update = DefaultUpdate
	}

	a.running = true
	a.dirty = true

	a.startPendingEffects()

	go a.pollEvents()

	var ticker *time.Ticker
	var ticks <-chan time.Time
	if a.tickRate > 0 {
		ticker = time.NewTicker(a.tickRate)
		defer ticker.Stop()
		ticks = ticker.C
	}

	for a.running {
		var msg Message
		select {
		case <-ctx.Done():
			a.running = false
			a.cancelTasks()
		case msg = <-a.messages:
			if a.update(a, msg) {
				a.dirty = true
			}
		case now := <-ticks:
			msg = TickMsg{Time: now}
			if a.update(a, msg) {
				a.dirty = true
			}
		}

		if !a.running {
			continue
		}

		if msg != nil {
			if a.flushQueueIfNeeded(msg) {
				a.dirty = true
			}
			if _, ok := msg.(InvalidateMsg); ok && a.invalidator != nil {
				a.invalidator.resetPending()
			}
		}

		if a.dirty {
			a.render()
			a.dirty = false
		}
	}

	return ctx.Err()
}

// DefaultUpdate handles input messages and widget commands.
func DefaultUpdate(app *App, msg Message) bool {
	if app == nil || app.screen == nil {
		return false
	}

	switch m := msg.(type) {
	case ResizeMsg:
		app.screen.Resize(m.Width, m.Height)
		if app.recorder != nil {
			_ = app.recorder.Resize(m.Width, m.Height)
		}
		return true
	case KeyMsg:
		if app.keyHandler != nil {
			var focused Widget
			if scope := app.screen.FocusScope(); scope != nil {
				focused = scope.Current()
			}
			if app.keyHandler.HandleKey(app, m, focused) {
				return true
			}
		}
		return app.dispatchMessage(msg)
	case QueueFlushMsg:
		return false
	case InvalidateMsg:
		return true
	default:
		return app.dispatchMessage(msg)
	}
}

func (a *App) dispatchMessage(msg Message) bool {
	if a == nil || a.screen == nil {
		return false
	}
	result := a.screen.HandleMessage(msg)
	dirty := result.Handled
	for _, cmd := range result.Commands {
		if a.handleCommand(cmd) {
			dirty = true
		}
	}
	return dirty
}

func (a *App) handleCommand(cmd Command) bool {
	switch c := cmd.(type) {
	case Quit:
		a.running = false
		a.cancelTasks()
		return false
	case Refresh:
		if a.screen != nil {
			a.screen.Buffer().MarkAllDirty()
		}
		return true
	case SendMsg:
		if c.Message != nil {
			a.Post(c.Message)
		}
		return false
	case Effect:
		a.runEffect(c)
		return false
	default:
		if a.commandHandler != nil {
			return a.commandHandler(cmd)
		}
		return false
	}
}

// ExecuteCommand runs a command through the app handler.
func (a *App) ExecuteCommand(cmd Command) bool {
	if a == nil {
		return false
	}
	return a.handleCommand(cmd)
}

func (a *App) pollEvents() {
	for a.running {
		ev := a.backend.PollEvent()
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case terminal.KeyEvent:
			a.Post(KeyMsg{
				Key:   e.Key,
				Rune:  e.Rune,
				Alt:   e.Alt,
				Ctrl:  e.Ctrl,
				Shift: e.Shift,
			})
		case terminal.ResizeEvent:
			a.Post(ResizeMsg{Width: e.Width, Height: e.Height})
		case terminal.MouseEvent:
			a.Post(MouseMsg{
				X:      e.X,
				Y:      e.Y,
				Button: MouseButton(e.Button),
				Action: MouseAction(e.Action),
				Alt:    e.Alt,
				Ctrl:   e.Ctrl,
				Shift:  e.Shift,
			})
		case terminal.PasteEvent:
			a.Post(PasteMsg{Text: e.Text})
		}
	}
}

func (a *App) render() {
	a.renderMu.Lock()
	defer a.renderMu.Unlock()

	if a.screen == nil {
		return
	}

	observer := a.renderObserver
	var stats RenderStats
	if observer != nil {
		stats.Frame = atomic.AddInt64(&a.renderFrame, 1)
		stats.Started = time.Now()
		stats.LayerCount = a.screen.LayerCount()
	}

	renderStart := time.Time{}
	if observer != nil {
		renderStart = time.Now()
	}
	a.screen.Render()
	if observer != nil {
		stats.RenderDuration = time.Since(renderStart)
	}
	buf := a.screen.Buffer()
	if observer != nil && buf != nil {
		w, h := buf.Size()
		stats.TotalCells = w * h
	}

	if buf.IsDirty() {
		dirtyCount := buf.DirtyCount()
		w, h := buf.Size()
		totalCells := w * h
		fullRedraw := dirtyCount > totalCells/2
		if observer != nil {
			stats.DirtyCells = dirtyCount
			stats.TotalCells = totalCells
			stats.DirtyRect = buf.DirtyRect()
			stats.FullRedraw = fullRedraw
		}

		rowWriter, hasRowWriter := a.backend.(backend.RowWriter)
		rectWriter, hasRectWriter := a.backend.(backend.RectWriter)
		cells := buf.Cells()
		flushedCells := 0
		flushStart := time.Time{}
		if observer != nil {
			flushStart = time.Now()
		}
		if fullRedraw {
			switch {
			case hasRectWriter:
				rectWriter.SetRect(0, 0, w, h, cells)
			case hasRowWriter:
				for y := 0; y < h; y++ {
					rowStart := y * w
					rowWriter.SetRow(y, 0, cells[rowStart:rowStart+w])
				}
			default:
				for y := 0; y < h; y++ {
					rowStart := y * w
					row := cells[rowStart : rowStart+w]
					for x, cell := range row {
						a.backend.SetContent(x, y, cell.Rune, nil, cell.Style)
					}
				}
			}
			flushedCells = totalCells
		} else {
			rect := buf.DirtyRect()
			rectArea := rect.Width * rect.Height
			useRect := hasRectWriter && rect.Width == w && rectArea > 0 && dirtyCount*2 >= rectArea
			if useRect {
				start := rect.Y * w
				end := start + rectArea
				rectWriter.SetRect(0, rect.Y, w, rect.Height, cells[start:end])
				flushedCells = rectArea
			} else if hasRowWriter && rectArea > 0 && dirtyCount*4 >= rectArea {
				buf.ForEachDirtySpan(func(y, startX, endX int) {
					rowStart := y * w
					rowWriter.SetRow(y, startX, cells[rowStart+startX:rowStart+endX])
					flushedCells += endX - startX
				})
			} else {
				buf.ForEachDirtyCell(func(x, y int, cell Cell) {
					a.backend.SetContent(x, y, cell.Rune, nil, cell.Style)
				})
				flushedCells = dirtyCount
			}
		}
		if observer != nil {
			stats.FlushDuration = time.Since(flushStart)
			stats.FlushedCells = flushedCells
		}
		if a.recorder != nil {
			if err := a.recorder.Frame(buf, time.Now()); err != nil {
				a.recorder = nil
			}
		}
		buf.ClearDirty()
	}

	a.backend.Show()
	if observer != nil {
		stats.Ended = time.Now()
		stats.TotalDuration = stats.Ended.Sub(stats.Started)
		observer.ObserveRender(stats)
	}
}

func (a *App) taskContext() context.Context {
	if a != nil && a.taskCtx != nil {
		return a.taskCtx
	}
	return context.Background()
}

func (a *App) cancelTasks() {
	if a == nil || a.taskCancel == nil {
		return
	}
	a.taskCancel()
}

func (a *App) runEffect(effect Effect) {
	if a == nil || effect.Run == nil {
		return
	}
	ctx := a.taskContext()
	post := a.tryPost
	go effect.Run(ctx, post)
}

func (a *App) startPendingEffects() {
	if a == nil {
		return
	}
	a.pendingMu.Lock()
	effects := a.pendingEffects
	a.pendingEffects = nil
	a.pendingMu.Unlock()
	for _, effect := range effects {
		a.runEffect(effect)
	}
}

func (a *App) flushQueueIfNeeded(msg Message) bool {
	if a == nil || a.stateQueue == nil {
		return false
	}
	if !shouldFlushQueue(a.flushPolicy, msg) {
		return false
	}
	if a.queueScheduler != nil {
		a.queueScheduler.resetPending()
	}
	return a.stateQueue.Flush() > 0
}
