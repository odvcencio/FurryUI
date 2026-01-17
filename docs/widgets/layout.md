# Layout Widgets

Use these widgets to compose structure. A complete demo is in
`examples/widgets/layout`.

## Grid

`Grid` divides space into rows and columns. Each child can span rows/columns.

API notes:
- `NewGrid(rows, cols)` sets the base grid.
- `Add(widget, row, col, rowSpan, colSpan)` positions children.
- `Gap` controls spacing between cells.
- GoDoc example: `ExampleGrid`.

Example:

```go
grid := widgets.NewGrid(2, 2)
grid.Gap = 1
grid.Add(widgets.NewLabel("Top"), 0, 0, 1, 2)
```

## Splitter

`Splitter` divides a region into two resizable panes.

API notes:
- `NewSplitter(first, second)` creates the container.
- `Orientation` chooses horizontal or vertical split.
- `Ratio` controls the split (0.0 - 1.0).
- GoDoc example: `ExampleSplitter`.

Example:

```go
split := widgets.NewSplitter(leftPane, rightPane)
split.Orientation = widgets.SplitHorizontal
split.Ratio = 0.6
```

## Stack

`Stack` overlays children in z-order.

API notes:
- `NewStack(children...)` creates the stack.
- Children are rendered in order; later items appear on top.
- GoDoc example: `ExampleStack`.

Example:

```go
stack := widgets.NewStack(background, overlay)
```

## ScrollView

`ScrollView` wraps content in a scrollable viewport.

API notes:
- `NewScrollView(content)` creates the container.
- `SetBehavior` configures scroll policies and page size.
- `ScrollBy`, `ScrollToStart`, and `ScrollToEnd` support programmatic control.
- Implement `scroll.VirtualSizer` / `scroll.VirtualIndexer` for fast virtual lists.
- GoDoc example: `ExampleScrollView`.

Example:

```go
scroll := widgets.NewScrollView(widgets.NewText(longText))
```

## Panel and Box

`Panel` draws a border and title around a child. `Box` fills a background.

API notes:
- `NewPanel(child)` returns a panel.
- `WithBorder(style)` enables a border.
- `SetTitle` labels the panel.
- `NewBox(child)` creates a background fill container.
- GoDoc example: `ExamplePanel`, `ExampleBox`.

Example:

```go
panel := widgets.NewPanel(content).WithBorder(backend.DefaultStyle())
panel.SetTitle("Details")
```
