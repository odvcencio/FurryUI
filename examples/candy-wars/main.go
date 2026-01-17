// Candy Wars - A FluffyUI Showcase Game
//
// Navigate the halls of Jefferson Middle School trading candy between
// locations. Buy low, sell high, and avoid getting caught by teachers!
//
// This game demonstrates FluffyUI's capabilities:
// - Reactive state management with Signals
// - Complex widget composition (Tables, Dialogs, Charts, Panels)
// - Keybinding system
// - Dynamic UI updates
// - Game loop with tick-based events
package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/examples/internal/demo"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
	"github.com/odvcencio/fluffy-ui/widgets"
)

func main() {
	// rand is auto-seeded in Go 1.20+

	game := NewGame()
	view := NewGameView(game)

	bundle, err := demo.NewApp(view, demo.Options{
		CommandHandler: func(cmd runtime.Command) bool {
			if _, ok := cmd.(runtime.Quit); ok {
				return true
			}
			return false
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init failed: %v\n", err)
		os.Exit(1)
	}

	// Game tick - price fluctuations and random events
	bundle.App.Every(2*time.Second, func(now time.Time) runtime.Message {
		if game.GameOver.Get() {
			return nil
		}
		game.Tick()
		return nil
	})

	if err := bundle.App.Run(context.Background()); err != nil && err != context.Canceled {
		fmt.Fprintf(os.Stderr, "app run failed: %v\n", err)
		os.Exit(1)
	}
}

// =============================================================================
// Game Data Types
// =============================================================================

type CandyType struct {
	Name     string
	MinPrice int
	MaxPrice int
	Emoji    string
}

var CandyTypes = []CandyType{
	{Name: "Gummy Bears", MinPrice: 1, MaxPrice: 8, Emoji: "[G]"},
	{Name: "Chocolate Bar", MinPrice: 5, MaxPrice: 25, Emoji: "[C]"},
	{Name: "Sour Straws", MinPrice: 2, MaxPrice: 15, Emoji: "[S]"},
	{Name: "Lollipops", MinPrice: 1, MaxPrice: 6, Emoji: "[L]"},
	{Name: "Jawbreakers", MinPrice: 3, MaxPrice: 20, Emoji: "[J]"},
	{Name: "Rare Import", MinPrice: 20, MaxPrice: 150, Emoji: "[R]"},
}

type Location struct {
	Name        string
	Description string
	RiskLevel   int // 1-5, higher = more teacher patrols
}

var Locations = []Location{
	{Name: "Cafeteria", Description: "Busy during lunch - good prices, moderate risk", RiskLevel: 3},
	{Name: "Gymnasium", Description: "Athletes pay premium for energy", RiskLevel: 2},
	{Name: "Library", Description: "Quiet trades, but librarian watches closely", RiskLevel: 4},
	{Name: "Playground", Description: "High demand, but teachers patrol often", RiskLevel: 5},
	{Name: "Art Room", Description: "Creative kids, unpredictable prices", RiskLevel: 2},
	{Name: "Music Hall", Description: "Band kids have allowance money", RiskLevel: 1},
}

type Inventory map[string]int

type MarketPrices map[string]int

// =============================================================================
// Game State
// =============================================================================

type Game struct {
	Cash          *state.Signal[int]
	Debt          *state.Signal[int]
	Day           *state.Signal[int]
	MaxDays       int
	Location      *state.Signal[int]
	Inventory     *state.Signal[Inventory]
	Capacity      int
	Prices        *state.Signal[MarketPrices]
	PriceHistory  *state.Signal[[]float64]
	Message       *state.Signal[string]
	GameOver      *state.Signal[bool]
	GameOverMsg   *state.Signal[string]
	ShowEvent     *state.Signal[bool]
	EventTitle    *state.Signal[string]
	EventMessage  *state.Signal[string]
	Heat          *state.Signal[int] // 0-100, teacher suspicion level
}

func NewGame() *Game {
	g := &Game{
		Cash:         state.NewSignal(100),
		Debt:         state.NewSignal(500),
		Day:          state.NewSignal(1),
		MaxDays:      30,
		Location:     state.NewSignal(0),
		Inventory:    state.NewSignal(make(Inventory)),
		Capacity:     100,
		Prices:       state.NewSignal(make(MarketPrices)),
		PriceHistory: state.NewSignal([]float64{100}),
		Message:      state.NewSignal("Welcome to Jefferson Middle School! Trade candy to pay off your debt."),
		GameOver:     state.NewSignal(false),
		GameOverMsg:  state.NewSignal(""),
		ShowEvent:    state.NewSignal(false),
		EventTitle:   state.NewSignal(""),
		EventMessage: state.NewSignal(""),
		Heat:         state.NewSignal(0),
	}

	// Set equality functions
	g.Cash.SetEqualFunc(state.EqualComparable[int])
	g.Debt.SetEqualFunc(state.EqualComparable[int])
	g.Day.SetEqualFunc(state.EqualComparable[int])
	g.Location.SetEqualFunc(state.EqualComparable[int])
	g.Message.SetEqualFunc(state.EqualComparable[string])
	g.GameOver.SetEqualFunc(state.EqualComparable[bool])
	g.GameOverMsg.SetEqualFunc(state.EqualComparable[string])
	g.ShowEvent.SetEqualFunc(state.EqualComparable[bool])
	g.EventTitle.SetEqualFunc(state.EqualComparable[string])
	g.EventMessage.SetEqualFunc(state.EqualComparable[string])
	g.Heat.SetEqualFunc(state.EqualComparable[int])

	g.GeneratePrices()
	return g
}

func (g *Game) GeneratePrices() {
	prices := make(MarketPrices)
	for _, candy := range CandyTypes {
		prices[candy.Name] = candy.MinPrice + rand.Intn(candy.MaxPrice-candy.MinPrice+1)
	}
	g.Prices.Set(prices)
}

func (g *Game) Tick() {
	// Fluctuate prices slightly
	prices := g.Prices.Get()
	newPrices := make(MarketPrices)
	for name, price := range prices {
		var candy CandyType
		for _, c := range CandyTypes {
			if c.Name == name {
				candy = c
				break
			}
		}
		delta := rand.Intn(5) - 2 // -2 to +2
		newPrice := price + delta
		if newPrice < candy.MinPrice {
			newPrice = candy.MinPrice
		}
		if newPrice > candy.MaxPrice {
			newPrice = candy.MaxPrice
		}
		newPrices[name] = newPrice
	}
	g.Prices.Set(newPrices)

	// Update price history (track total inventory value)
	history := g.PriceHistory.Get()
	totalValue := float64(g.Cash.Get())
	inv := g.Inventory.Get()
	for name, qty := range inv {
		if price, ok := newPrices[name]; ok {
			totalValue += float64(qty * price)
		}
	}
	history = append(history, totalValue)
	if len(history) > 30 {
		history = history[len(history)-30:]
	}
	g.PriceHistory.Set(history)

	// Reduce heat over time
	heat := g.Heat.Get()
	if heat > 0 {
		g.Heat.Set(heat - 5)
	}
}

func (g *Game) Travel(locationIndex int) {
	if g.GameOver.Get() {
		return
	}
	if locationIndex < 0 || locationIndex >= len(Locations) {
		return
	}

	g.Location.Set(locationIndex)
	g.Day.Update(func(d int) int { return d + 1 })
	g.GeneratePrices()

	loc := Locations[locationIndex]
	g.Message.Set(fmt.Sprintf("You arrived at the %s. %s", loc.Name, loc.Description))

	// Check for random events
	g.CheckRandomEvent()

	// Check win/lose conditions
	g.CheckEndConditions()
}

func (g *Game) CheckRandomEvent() {
	loc := Locations[g.Location.Get()]
	heat := g.Heat.Get()

	// Base chance + location risk + heat
	catchChance := loc.RiskLevel*5 + heat/2

	roll := rand.Intn(100)

	if roll < catchChance {
		// Caught!
		g.TriggerEvent("Busted!", g.getCaughtEvent())
	} else if roll < 20 {
		// Good event
		g.TriggerGoodEvent()
	} else if roll < 30 {
		// Price spike event
		g.TriggerPriceEvent()
	}
}

func (g *Game) getCaughtEvent() string {
	events := []struct {
		msg    string
		action func()
	}{
		{
			msg: "A teacher spotted your candy stash!\nThey confiscated half your inventory.",
			action: func() {
				inv := g.Inventory.Get()
				newInv := make(Inventory)
				for name, qty := range inv {
					newInv[name] = qty / 2
				}
				g.Inventory.Set(newInv)
			},
		},
		{
			msg: "The principal caught you trading!\nYou paid a $50 'fine' to keep quiet.",
			action: func() {
				cash := g.Cash.Get()
				if cash >= 50 {
					g.Cash.Set(cash - 50)
				} else {
					g.Cash.Set(0)
				}
			},
		},
		{
			msg: "Hall monitor shakedown!\nLost $25 and some candy.",
			action: func() {
				g.Cash.Update(func(c int) int {
					if c >= 25 {
						return c - 25
					}
					return 0
				})
				inv := g.Inventory.Get()
				newInv := make(Inventory)
				for name, qty := range inv {
					if qty > 2 {
						newInv[name] = qty - 2
					} else {
						newInv[name] = 0
					}
				}
				g.Inventory.Set(newInv)
			},
		},
	}

	event := events[rand.Intn(len(events))]
	event.action()
	g.Heat.Set(0) // Reset heat after getting caught
	return event.msg
}

func (g *Game) TriggerGoodEvent() {
	events := []struct {
		title string
		msg   string
		action func()
	}{
		{
			title: "Lucky Find!",
			msg:   "You found $20 in the hallway!",
			action: func() {
				g.Cash.Update(func(c int) int { return c + 20 })
			},
		},
		{
			title: "Generous Kid",
			msg:   "A kid gave you 5 Gummy Bears for helping with homework!",
			action: func() {
				inv := g.Inventory.Get()
				newInv := make(Inventory)
				for k, v := range inv {
					newInv[k] = v
				}
				newInv["Gummy Bears"] += 5
				g.Inventory.Set(newInv)
			},
		},
		{
			title: "Debt Forgiveness",
			msg:   "The kid you owe money forgot about $50 of your debt!",
			action: func() {
				g.Debt.Update(func(d int) int {
					if d >= 50 {
						return d - 50
					}
					return 0
				})
			},
		},
	}

	event := events[rand.Intn(len(events))]
	event.action()
	g.TriggerEvent(event.title, event.msg)
}

func (g *Game) TriggerPriceEvent() {
	candy := CandyTypes[rand.Intn(len(CandyTypes))]
	prices := g.Prices.Get()
	newPrices := make(MarketPrices)
	for k, v := range prices {
		newPrices[k] = v
	}

	if rand.Intn(2) == 0 {
		// Price crash
		newPrices[candy.Name] = candy.MinPrice
		g.Prices.Set(newPrices)
		g.TriggerEvent("Market Crash!", fmt.Sprintf("%s prices have crashed!\nBuying opportunity?", candy.Name))
	} else {
		// Price spike
		newPrices[candy.Name] = candy.MaxPrice
		g.Prices.Set(newPrices)
		g.TriggerEvent("Price Spike!", fmt.Sprintf("%s is in high demand!\nSell now for maximum profit!", candy.Name))
	}
}

func (g *Game) TriggerEvent(title, msg string) {
	g.EventTitle.Set(title)
	g.EventMessage.Set(msg)
	g.ShowEvent.Set(true)
}

func (g *Game) DismissEvent() {
	g.ShowEvent.Set(false)
}

func (g *Game) CheckEndConditions() {
	day := g.Day.Get()
	cash := g.Cash.Get()
	debt := g.Debt.Get()

	// Calculate total worth
	inv := g.Inventory.Get()
	prices := g.Prices.Get()
	totalWorth := cash
	for name, qty := range inv {
		if price, ok := prices[name]; ok {
			totalWorth += qty * price
		}
	}

	if totalWorth >= debt && debt > 0 {
		// Can pay off debt!
		g.GameOver.Set(true)
		g.GameOverMsg.Set(fmt.Sprintf(
			"Congratulations!\n\nYou paid off your $%d debt in %d days!\n\nFinal worth: $%d\nProfit: $%d\n\nYou're the Candy King of Jefferson Middle!",
			debt, day, totalWorth, totalWorth-debt,
		))
		return
	}

	if day >= g.MaxDays {
		if totalWorth >= debt {
			g.GameOver.Set(true)
			g.GameOverMsg.Set(fmt.Sprintf(
				"Time's Up - But You Won!\n\nYou made $%d and paid your debt!\n\nCongratulations, Candy Trader!",
				totalWorth,
			))
		} else {
			g.GameOver.Set(true)
			g.GameOverMsg.Set(fmt.Sprintf(
				"Game Over!\n\n30 days have passed.\nYou still owe $%d.\n\nThe candy mafia is not pleased...\n\nFinal worth: $%d",
				debt-totalWorth, totalWorth,
			))
		}
	}
}

func (g *Game) Buy(candyName string, qty int) bool {
	if g.GameOver.Get() {
		return false
	}

	prices := g.Prices.Get()
	price, ok := prices[candyName]
	if !ok {
		return false
	}

	totalCost := price * qty
	cash := g.Cash.Get()
	if totalCost > cash {
		g.Message.Set("Not enough cash!")
		return false
	}

	inv := g.Inventory.Get()
	currentQty := 0
	for _, q := range inv {
		currentQty += q
	}
	if currentQty+qty > g.Capacity {
		g.Message.Set(fmt.Sprintf("Not enough space! (Capacity: %d/%d)", currentQty, g.Capacity))
		return false
	}

	// Execute trade
	g.Cash.Set(cash - totalCost)
	newInv := make(Inventory)
	for k, v := range inv {
		newInv[k] = v
	}
	newInv[candyName] += qty
	g.Inventory.Set(newInv)

	// Increase heat
	g.Heat.Update(func(h int) int {
		newHeat := h + qty
		if newHeat > 100 {
			return 100
		}
		return newHeat
	})

	g.Message.Set(fmt.Sprintf("Bought %d %s for $%d", qty, candyName, totalCost))
	return true
}

func (g *Game) Sell(candyName string, qty int) bool {
	if g.GameOver.Get() {
		return false
	}

	inv := g.Inventory.Get()
	owned := inv[candyName]
	if qty > owned {
		g.Message.Set("You don't have that much!")
		return false
	}

	prices := g.Prices.Get()
	price, ok := prices[candyName]
	if !ok {
		return false
	}

	totalValue := price * qty

	// Execute trade
	g.Cash.Update(func(c int) int { return c + totalValue })
	newInv := make(Inventory)
	for k, v := range inv {
		newInv[k] = v
	}
	newInv[candyName] -= qty
	if newInv[candyName] <= 0 {
		delete(newInv, candyName)
	}
	g.Inventory.Set(newInv)

	// Increase heat (selling is riskier)
	g.Heat.Update(func(h int) int {
		newHeat := h + qty*2
		if newHeat > 100 {
			return 100
		}
		return newHeat
	})

	g.Message.Set(fmt.Sprintf("Sold %d %s for $%d", qty, candyName, totalValue))
	return true
}

func (g *Game) PayDebt(amount int) bool {
	cash := g.Cash.Get()
	debt := g.Debt.Get()

	if amount > cash {
		amount = cash
	}
	if amount > debt {
		amount = debt
	}
	if amount <= 0 {
		return false
	}

	g.Cash.Set(cash - amount)
	g.Debt.Set(debt - amount)
	g.Message.Set(fmt.Sprintf("Paid $%d towards debt. Remaining: $%d", amount, debt-amount))

	if debt-amount <= 0 {
		g.GameOver.Set(true)
		g.GameOverMsg.Set("You paid off all your debt!\n\nYou win!")
	}

	return true
}

func (g *Game) InventoryCount() int {
	inv := g.Inventory.Get()
	count := 0
	for _, q := range inv {
		count += q
	}
	return count
}

// =============================================================================
// Game View
// =============================================================================

type GameView struct {
	widgets.Component
	game *Game

	// UI elements
	header       *widgets.Label
	statusPanel  *widgets.Panel
	marketTable  *widgets.Table
	locationList *widgets.List[Location]
	messageLabel *widgets.Label
	inventoryLbl *widgets.Label
	heatGauge    *widgets.Progress
	sparkline    *widgets.Sparkline

	// Trade dialog state
	showTrade     bool
	tradeCandy    string
	tradeIsBuy    bool
	tradeQty      int
	tradeInput    *widgets.Input

	// Layout
	focusIndex   int
	style        backend.Style
	dimStyle     backend.Style
	accentStyle  backend.Style
	successStyle backend.Style
	dangerStyle  backend.Style
	warningStyle backend.Style
}

func NewGameView(game *Game) *GameView {
	v := &GameView{
		game:         game,
		style:        backend.DefaultStyle(),
		dimStyle:     backend.DefaultStyle().Dim(true),
		accentStyle:  backend.DefaultStyle().Bold(true),
		successStyle: backend.DefaultStyle().Foreground(backend.ColorGreen).Bold(true),
		dangerStyle:  backend.DefaultStyle().Foreground(backend.ColorRed).Bold(true),
		warningStyle: backend.DefaultStyle().Foreground(backend.ColorYellow),
	}

	v.header = widgets.NewLabel("CANDY WARS - Jefferson Middle School").WithStyle(backend.DefaultStyle().Bold(true).Foreground(backend.ColorYellow))

	// Market table
	v.marketTable = widgets.NewTable(
		widgets.TableColumn{Title: "Candy", Width: 15},
		widgets.TableColumn{Title: "Price", Width: 8},
		widgets.TableColumn{Title: "Owned", Width: 8},
	)
	v.updateMarketTable()

	// Location list
	adapter := widgets.NewSliceAdapter(Locations, func(loc Location, index int, selected bool, ctx runtime.RenderContext) {
		style := v.style
		if selected {
			style = style.Reverse(true)
		}
		riskStars := ""
		for i := 0; i < loc.RiskLevel; i++ {
			riskStars += "*"
		}
		line := fmt.Sprintf("%-12s [%s]", loc.Name, riskStars)
		line = truncPad(line, ctx.Bounds.Width)
		ctx.Buffer.SetString(ctx.Bounds.X, ctx.Bounds.Y, line, style)
	})
	v.locationList = widgets.NewList(adapter)
	v.locationList.OnSelect(func(index int, loc Location) {
		v.game.Travel(index)
		v.refresh()
	})

	v.messageLabel = widgets.NewLabel("")
	v.inventoryLbl = widgets.NewLabel("")

	v.heatGauge = widgets.NewProgress()
	v.heatGauge.Max = 100
	v.heatGauge.Label = "Heat"

	v.sparkline = widgets.NewSparkline(game.PriceHistory)

	v.tradeInput = widgets.NewInput()
	v.tradeInput.SetPlaceholder("Quantity")

	v.refresh()
	return v
}

func (v *GameView) Mount() {
	v.Observe(v.game.Cash, v.refresh)
	v.Observe(v.game.Debt, v.refresh)
	v.Observe(v.game.Day, v.refresh)
	v.Observe(v.game.Prices, v.refresh)
	v.Observe(v.game.Inventory, v.refresh)
	v.Observe(v.game.Message, v.refresh)
	v.Observe(v.game.Heat, v.refresh)
	v.Observe(v.game.Location, v.refresh)
	v.Observe(v.game.ShowEvent, v.refresh)
	v.Observe(v.game.GameOver, v.refresh)
	v.refresh()
}

func (v *GameView) Unmount() {
	v.Subs.Clear()
}

func (v *GameView) refresh() {
	v.updateMarketTable()
	v.messageLabel.SetText(v.game.Message.Get())

	inv := v.game.Inventory.Get()
	invText := fmt.Sprintf("Backpack: %d/%d", v.game.InventoryCount(), v.game.Capacity)
	if len(inv) > 0 {
		invText += " |"
		for name, qty := range inv {
			if qty > 0 {
				invText += fmt.Sprintf(" %s:%d", shortName(name), qty)
			}
		}
	}
	v.inventoryLbl.SetText(invText)

	heat := v.game.Heat.Get()
	v.heatGauge.Value = float64(heat)

	v.Invalidate()
}

func (v *GameView) updateMarketTable() {
	prices := v.game.Prices.Get()
	inv := v.game.Inventory.Get()

	rows := make([][]string, len(CandyTypes))
	for i, candy := range CandyTypes {
		price := prices[candy.Name]
		owned := inv[candy.Name]
		rows[i] = []string{
			candy.Emoji + " " + candy.Name,
			fmt.Sprintf("$%d", price),
			fmt.Sprintf("%d", owned),
		}
	}
	v.marketTable.SetRows(rows)
}

func (v *GameView) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

func (v *GameView) Layout(bounds runtime.Rect) {
	v.Component.Layout(bounds)

	// Header row
	v.header.Layout(runtime.Rect{X: bounds.X, Y: bounds.Y, Width: bounds.Width, Height: 1})

	// Status bar (Day, Cash, Debt, Heat)
	y := bounds.Y + 2

	// Left side: Market (55% width)
	leftWidth := bounds.Width * 55 / 100
	rightWidth := bounds.Width - leftWidth - 1

	// Market table
	marketHeight := len(CandyTypes) + 2
	v.marketTable.Layout(runtime.Rect{X: bounds.X, Y: y, Width: leftWidth, Height: marketHeight})

	// Location list (right side)
	v.locationList.Layout(runtime.Rect{X: bounds.X + leftWidth + 1, Y: y, Width: rightWidth, Height: len(Locations) + 1})

	y += marketHeight + 1

	// Sparkline
	v.sparkline.Layout(runtime.Rect{X: bounds.X, Y: y, Width: leftWidth, Height: 1})

	// Heat gauge
	v.heatGauge.Layout(runtime.Rect{X: bounds.X + leftWidth + 1, Y: y, Width: rightWidth, Height: 1})

	y += 2

	// Inventory
	v.inventoryLbl.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 1})

	y += 2

	// Message
	v.messageLabel.Layout(runtime.Rect{X: bounds.X, Y: y, Width: bounds.Width, Height: 2})

	// Trade input (hidden unless trading)
	if v.showTrade {
		v.tradeInput.Layout(runtime.Rect{X: bounds.X + 20, Y: bounds.Height/2 + 2, Width: 20, Height: 1})
	}
}

func (v *GameView) Render(ctx runtime.RenderContext) {
	if ctx.Buffer == nil {
		return
	}
	bounds := v.Bounds()
	ctx.Clear(v.style)

	// Header
	v.header.Render(ctx)

	// Status bar with colored segments
	y := bounds.Y + 1
	loc := Locations[v.game.Location.Get()]
	x := bounds.X

	// Day
	dayText := fmt.Sprintf("Day %d/%d", v.game.Day.Get(), v.game.MaxDays)
	ctx.Buffer.SetString(x, y, dayText, v.accentStyle)
	x += len(dayText) + 1

	ctx.Buffer.SetString(x, y, "|", v.dimStyle)
	x += 2

	// Cash (green)
	cashText := fmt.Sprintf("Cash: $%d", v.game.Cash.Get())
	ctx.Buffer.SetString(x, y, cashText, v.successStyle)
	x += len(cashText) + 1

	ctx.Buffer.SetString(x, y, "|", v.dimStyle)
	x += 2

	// Debt (red)
	debtText := fmt.Sprintf("Debt: $%d", v.game.Debt.Get())
	ctx.Buffer.SetString(x, y, debtText, v.dangerStyle)
	x += len(debtText) + 1

	ctx.Buffer.SetString(x, y, "|", v.dimStyle)
	x += 2

	// Location
	locText := fmt.Sprintf("Location: %s", loc.Name)
	ctx.Buffer.SetString(x, y, locText, v.accentStyle)

	// Market table
	v.marketTable.Render(ctx)

	// Location list with header
	listBounds := v.locationList.Bounds()
	ctx.Buffer.SetString(listBounds.X, listBounds.Y-1, "Travel To:", v.accentStyle)
	v.locationList.Render(ctx)

	// Sparkline with label
	sparkBounds := v.sparkline.Bounds()
	ctx.Buffer.SetString(sparkBounds.X, sparkBounds.Y, "Net Worth: ", v.dimStyle)
	v.sparkline.Render(ctx)

	// Heat gauge
	v.heatGauge.Render(ctx)

	// Inventory
	v.inventoryLbl.Render(ctx)

	// Message
	v.messageLabel.Render(ctx)

	// Help text
	helpY := bounds.Y + bounds.Height - 2
	help := "[B]uy  [S]ell  [P]ay Debt  [1-6]Travel  [Q]uit"
	ctx.Buffer.SetString(bounds.X, helpY, help, v.dimStyle)

	// Trade dialog
	if v.showTrade {
		v.renderTradeDialog(ctx)
	}

	// Event dialog
	if v.game.ShowEvent.Get() {
		v.renderEventDialog(ctx)
	}

	// Game over dialog
	if v.game.GameOver.Get() {
		v.renderGameOverDialog(ctx)
	}
}

func (v *GameView) renderTradeDialog(ctx runtime.RenderContext) {
	bounds := v.Bounds()
	dialogW := 40
	dialogH := 8
	x := (bounds.Width - dialogW) / 2
	y := (bounds.Height - dialogH) / 2

	rect := runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	action := "Buy"
	if !v.tradeIsBuy {
		action = "Sell"
	}
	title := fmt.Sprintf(" %s %s ", action, v.tradeCandy)
	ctx.Buffer.SetString(x+2, y, title, v.accentStyle)

	prices := v.game.Prices.Get()
	price := prices[v.tradeCandy]
	inv := v.game.Inventory.Get()
	owned := inv[v.tradeCandy]

	ctx.Buffer.SetString(x+2, y+2, fmt.Sprintf("Price: $%d each", price), v.style)
	ctx.Buffer.SetString(x+2, y+3, fmt.Sprintf("You own: %d", owned), v.style)
	ctx.Buffer.SetString(x+2, y+4, fmt.Sprintf("Cash: $%d", v.game.Cash.Get()), v.style)

	ctx.Buffer.SetString(x+2, y+6, "Qty: ", v.style)
	v.tradeInput.Render(ctx)

	ctx.Buffer.SetString(x+dialogW-15, y+6, "[Enter] [Esc]", v.dimStyle)
}

func (v *GameView) renderEventDialog(ctx runtime.RenderContext) {
	bounds := v.Bounds()
	dialogW := 50
	dialogH := 10
	x := (bounds.Width - dialogW) / 2
	y := (bounds.Height - dialogH) / 2

	rect := runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	title := " " + v.game.EventTitle.Get() + " "
	ctx.Buffer.SetString(x+2, y, title, v.accentStyle.Reverse(true))

	msg := v.game.EventMessage.Get()
	lines := splitLines(msg, dialogW-4)
	for i, line := range lines {
		if i < dialogH-3 {
			ctx.Buffer.SetString(x+2, y+2+i, line, v.style)
		}
	}

	ctx.Buffer.SetString(x+2, y+dialogH-2, "[Press any key to continue]", v.dimStyle)
}

func (v *GameView) renderGameOverDialog(ctx runtime.RenderContext) {
	bounds := v.Bounds()
	dialogW := 50
	dialogH := 14
	x := (bounds.Width - dialogW) / 2
	y := (bounds.Height - dialogH) / 2

	rect := runtime.Rect{X: x, Y: y, Width: dialogW, Height: dialogH}
	ctx.Buffer.Fill(rect, ' ', v.style)
	ctx.Buffer.DrawBox(rect, v.accentStyle)

	ctx.Buffer.SetString(x+2, y, " GAME OVER ", v.accentStyle.Reverse(true))

	msg := v.game.GameOverMsg.Get()
	lines := splitLines(msg, dialogW-4)
	for i, line := range lines {
		if i < dialogH-3 {
			ctx.Buffer.SetString(x+2, y+2+i, line, v.style)
		}
	}

	ctx.Buffer.SetString(x+2, y+dialogH-2, "[Press Q to quit]", v.dimStyle)
}

func (v *GameView) HandleMessage(msg runtime.Message) runtime.HandleResult {
	// Handle game over state
	if v.game.GameOver.Get() {
		if key, ok := msg.(runtime.KeyMsg); ok {
			if key.Rune == 'q' || key.Rune == 'Q' {
				return runtime.WithCommand(runtime.Quit{})
			}
		}
		return runtime.Handled()
	}

	// Handle event dialog
	if v.game.ShowEvent.Get() {
		if _, ok := msg.(runtime.KeyMsg); ok {
			v.game.DismissEvent()
			v.Invalidate()
			return runtime.Handled()
		}
	}

	// Handle trade dialog
	if v.showTrade {
		return v.handleTradeInput(msg)
	}

	// Normal gameplay
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		// Try child widgets
		if result := v.marketTable.HandleMessage(msg); result.Handled {
			return result
		}
		if result := v.locationList.HandleMessage(msg); result.Handled {
			return result
		}
		return runtime.Unhandled()
	}

	switch key.Rune {
	case 'q', 'Q':
		return runtime.WithCommand(runtime.Quit{})

	case 'b', 'B':
		v.startTrade(true)
		return runtime.Handled()

	case 's', 'S':
		v.startTrade(false)
		return runtime.Handled()

	case 'p', 'P':
		v.game.PayDebt(v.game.Cash.Get())
		v.refresh()
		return runtime.Handled()

	case '1', '2', '3', '4', '5', '6':
		idx := int(key.Rune - '1')
		if idx >= 0 && idx < len(Locations) {
			v.game.Travel(idx)
			v.refresh()
		}
		return runtime.Handled()
	}

	// Pass to market table for navigation
	if result := v.marketTable.HandleMessage(msg); result.Handled {
		return result
	}

	return runtime.Unhandled()
}

func (v *GameView) startTrade(isBuy bool) {
	// Get selected candy from table
	// For simplicity, use the first candy or let user select
	v.showTrade = true
	v.tradeIsBuy = isBuy
	v.tradeCandy = CandyTypes[0].Name // Default to first
	v.tradeQty = 1
	v.tradeInput.Clear()
	v.tradeInput.SetPlaceholder("1")
	v.Invalidate()
}

func (v *GameView) handleTradeInput(msg runtime.Message) runtime.HandleResult {
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return v.tradeInput.HandleMessage(msg)
	}

	switch key.Key {
	case 27: // Escape
		v.showTrade = false
		v.Invalidate()
		return runtime.Handled()
	}

	switch key.Rune {
	case 13, 10: // Enter
		// Parse quantity
		qtyStr := v.tradeInput.Text()
		if qtyStr == "" {
			qtyStr = "1"
		}
		qty, err := strconv.Atoi(qtyStr)
		if err != nil || qty <= 0 {
			qty = 1
		}

		if v.tradeIsBuy {
			v.game.Buy(v.tradeCandy, qty)
		} else {
			v.game.Sell(v.tradeCandy, qty)
		}

		v.showTrade = false
		v.refresh()
		return runtime.Handled()

	case '1', '2', '3', '4', '5', '6':
		// Quick select candy type
		idx := int(key.Rune - '1')
		if idx >= 0 && idx < len(CandyTypes) {
			v.tradeCandy = CandyTypes[idx].Name
			v.Invalidate()
		}
		return runtime.Handled()

	default:
		// Pass to input
		return v.tradeInput.HandleMessage(msg)
	}
}

func (v *GameView) ChildWidgets() []runtime.Widget {
	return []runtime.Widget{
		v.header,
		v.marketTable,
		v.locationList,
		v.messageLabel,
		v.inventoryLbl,
		v.heatGauge,
		v.sparkline,
	}
}

// =============================================================================
// Helpers
// =============================================================================

func truncPad(s string, width int) string {
	if len(s) > width {
		return s[:width]
	}
	for len(s) < width {
		s += " "
	}
	return s
}

func shortName(name string) string {
	if len(name) > 6 {
		return name[:6]
	}
	return name
}

func splitLines(s string, maxWidth int) []string {
	var lines []string
	current := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
			continue
		}
		current += string(r)
		if len(current) >= maxWidth {
			lines = append(lines, current)
			current = ""
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
