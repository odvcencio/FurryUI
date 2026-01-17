# Performance

FluffyUI is designed for efficient terminal rendering, but large apps still
benefit from a few best practices.

## Prefer signals over polling

Use `state.Signal` and `state.Computed` to drive rendering. When values change,
call `Invalidate` once so the runtime can refresh on the next tick.

## Use ScrollView for large content

Wrap long content in `ScrollView` and implement `scroll.VirtualContent` when
possible to avoid rendering off-screen rows.
For faster indexing, implement `scroll.VirtualSizer` and `scroll.VirtualIndexer`
to avoid O(n) scans on every scroll update.

## Avoid full-screen redraws

Widgets should draw only the content they own. The runtime tracks dirty cells
and flushes minimal regions.

## Keep allocations low in Render

Avoid building large strings inside hot render loops. Precompute labels or cache
layout where possible.

## Render profiling

Capture render timings and dirty stats by wiring a render observer:

```go
sampler := runtime.NewRenderSampler(120)
app := runtime.NewApp(runtime.AppConfig{
    RenderObserver: sampler,
})

summary := sampler.Summary()
```

`RenderStats` includes render/flush durations, dirty cell counts, and the dirty
bounding box for each frame.

## Simulation backend

Use the `backend/sim` package in tests to verify rendering logic without a real
terminal.
