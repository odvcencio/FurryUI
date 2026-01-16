package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/odvcencio/furry-ui/backend"
	"github.com/odvcencio/furry-ui/state"
	"github.com/odvcencio/furry-ui/terminal"
)

// UpdateFunc handles a message and returns true if a render is needed.
type UpdateFunc func(app *App, msg Message) bool

// CommandHandler handles commands emitted by widgets.
// Return true if the command requires a render.
type CommandHandler func(cmd Command) bool

// AppConfig configures a runtime App.
type AppConfig struct {
	Backend        backend.Backend
	Root           Widget
	Update         UpdateFunc
	CommandHandler CommandHandler
	MessageBuffer  int
	TickRate       time.Duration
	StateQueue     *state.Queue
	FlushPolicy    QueueFlushPolicy
}

// App runs a widget tree against a terminal backend.
type App struct {
	backend        backend.Backend
	screen         *Screen
	root           Widget
	update         UpdateFunc
	commandHandler CommandHandler
	messages       chan Message
	tickRate       time.Duration
	stateQueue     *state.Queue
	queueScheduler *QueueScheduler
	flushPolicy    QueueFlushPolicy
	invalidator    *Invalidator
	taskCtx        context.Context
	taskCancel     context.CancelFunc
	pendingMu      sync.Mutex
	pendingEffects []Effect

	running  bool
	dirty    bool
	renderMu sync.Mutex
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
		backend:        cfg.Backend,
		root:           cfg.Root,
		update:         cfg.Update,
		commandHandler: cfg.CommandHandler,
		messages:       make(chan Message, bufferSize),
		tickRate:       cfg.TickRate,
		stateQueue:     queue,
		flushPolicy:    policy,
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
	if a.root != nil {
		a.screen.SetRoot(a.root)
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
		return true
	case QueueFlushMsg:
		return false
	case InvalidateMsg:
		return true
	default:
		result := app.screen.HandleMessage(msg)
		dirty := result.Handled
		for _, cmd := range result.Commands {
			if app.handleCommand(cmd) {
				dirty = true
			}
		}
		return dirty
	}
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

	a.screen.Render()
	buf := a.screen.Buffer()

	if buf.IsDirty() {
		dirtyCount := buf.DirtyCount()
		w, h := buf.Size()
		totalCells := w * h

		if dirtyCount > totalCells/2 {
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					cell := buf.Get(x, y)
					a.backend.SetContent(x, y, cell.Rune, nil, cell.Style)
				}
			}
		} else {
			buf.ForEachDirtyCell(func(x, y int, cell Cell) {
				a.backend.SetContent(x, y, cell.Rune, nil, cell.Style)
			})
		}
		buf.ClearDirty()
	}

	a.backend.Show()
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
