# scrollview

A [Bubble Tea](https://github.com/charmbracelet/bubbletea) component that extends [`charm.land/bubbles/v2/viewport`](https://pkg.go.dev/charm.land/bubbles/v2/viewport) with a 1-column scrollbar.

## Features

- **Embedded viewport** — `scrollview.Model` embeds `viewport.Model`, so all standard viewport methods (`SetHeight`, `Height`, `YOffset`, `SetYOffset`, `Init`, and keyboard/mouse handling via `Update`) are available directly without any wrappers.
- **Arrow indicators** — `▲`/`▼` always appear at the top and bottom cells of the scrollbar column. They render in **white** when more content is available in that direction, and in **gray** when already at the limit.
- **Single-cell thumb** — a `▒` block tracks the scroll position in the inner cells between the two arrows.
- **Styleable** — every element (`Track`, `Thumb`, `MoreAbove`, `NoMoreAbove`, `MoreBelow`, `NoMoreBelow`) is a `lipgloss.Style` with a `SetString`-embedded character, fully overridable at runtime.
- **Left or right** — the `Position` field controls which side the scrollbar column appears on (`scrollview.Left` or `scrollview.Right`, default is right).

## Usage

```go
vp := scrollview.New()
vp.SetWidth(60)   // total width, including the 1-column scrollbar
vp.SetHeight(20)
vp.SetContent(someMultiLineString)

// in your model's Update
vp, cmd = vp.Update(msg)

// in your model's View
return vp.View()  // returns a string
```

### Custom styles

```go
vp.Styles.Thumb = vp.Styles.Thumb.SetString("█").Foreground(lipgloss.Color("#ff8800"))
vp.Styles.MoreAbove = vp.Styles.MoreAbove.SetString("↑")
vp.Styles.MoreBelow = vp.Styles.MoreBelow.SetString("↓")
```

## ScrollTo

`ScrollTo(line int)` scrolls so that `line` is visible. When scrolling **upward** the line lands at the **bottom** of the viewport, maximising visible context above it — useful when a non-focusable heading precedes the first focusable row.

## Example

```
go run ./examples/
```
