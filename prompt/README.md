# prompt

A configurable inline prompt bubble for [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) that accepts one of a set of single-character keys and echoes the chosen key inline. Unrecognized keys briefly flash in place of the cursor instead of being silently ignored.

![prompt example](./example.gif)

## Install

```sh
go get github.com/gzigzigzeo/bubbles/prompt
```

## Quick start

The common case — a plain yes/no confirmation:

```go
p, err := prompt.New("Deploy now?",
    prompt.WithYesNoDefaultYes(),
    prompt.WithSuccessStyles(),
)

// In your model's Init:
func (m Model) Init() tea.Cmd {
    return p.Focus()
}

// In your model's Update:
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg.(type) {
    case prompt.YesMsg:
        // user confirmed
    case prompt.NoMsg:
        // user declined
    }
    _, cmd := p.Update(msg)
    return m, cmd
}

// In your model's View:
func (m Model) View() tea.View {
    return p.View()
}
```

For anything other than yes/no, `WithOption` registers one key at a time
with its own message:

```go
p, err := prompt.New("Overwrite, skip, or abort?",
    prompt.WithOption('o', OverwriteMsg{}),
    prompt.WithOption('s', SkipMsg{}),
    prompt.WithOption('a', AbortMsg{}),
)
```

`msg` must be non-nil — `New` returns an error if any registered key has a
nil `Msg`, if two options register the same key, or if `WithDefault` names a
key that wasn't registered.

## Styles

```go
type Styles struct {
    Container       lipgloss.Style // total width is Container.GetWidth()
    Icon            lipgloss.Style // use SetString + Width for a fixed-width glyph column
    Label           lipgloss.Style // question text color
    CursorStyle     lipgloss.Style // cursor block style
    CursorTextStyle lipgloss.Style // cursor character style when blinking
    Echo            lipgloss.Style // echoed answer style
    Invalid         lipgloss.Style // invalid-key hint style
}
```

### Built-in presets

| Function            | Color   | Icon |
|---------------------|---------|------|
| `NewWarnStyles()`   | Yellow  | ⚠    |
| `NewErrorStyles()`  | Orange  | !    |
| `NewSuccessStyles()`| Green   | ✓    |
| `NewInfoStyles()`   | Default | i    |

Override the container width after calling a preset:

```go
s := prompt.NewWarnStyles()
s.Container = s.Container.Width(60)
p.SetStyles(s)
```

## API reference

| Method | Description |
|--------|-------------|
| `New(question string, opts ...Option) (*PromptModel, error)` | Create a prompt from the given `Option`s; errors on a duplicate key, a nil `Msg`, or an unregistered default key |
| `WithYesNo()` | Registers `'y'` (`YesMsg`) and `'n'` (`NoMsg`), no default |
| `WithYesNoDefaultYes()` | Like `WithYesNo`, but registers `'Y'` (uppercase, the default) instead of `'y'` |
| `WithYesNoDefaultNo()` | Like `WithYesNo`, but registers `'N'` (uppercase, the default) instead of `'n'` |
| `WithOption(key rune, msg tea.Msg)` | Register one key with its own message (`msg` must be non-nil) |
| `WithDefault(key rune)` | Make Enter emit this key's registered Msg, as if pressed directly (general-purpose; must name an already-registered key) |
| `WithStyles(Styles)` | Apply style configuration at construction |
| `WithWarnStyles()` / `WithErrorStyles()` / `WithSuccessStyles()` / `WithInfoStyles()` | Shorthand for `WithStyles(NewXStyles())` |
| `WithAcceptByEnter(accept bool)` | Enable/disable Enter triggering the default (on by default) |
| `WithInvalidKeyDuration(time.Duration)` | How long the invalid-key hint stays visible (default 600ms) |
| `WithCursor(cursor.Model)` | Seed the initial cursor instead of `cursor.New()` |
| `SetStyles(Styles)` | Apply style configuration at runtime (e.g. switching themes on a focused prompt) |
| `Init() tea.Cmd` | Starts cursor blinking (satisfies `tea.Model`) |
| `Focus() tea.Cmd` | Focus, reset answer, start cursor |
| `Blur()` | Unfocus, stop cursor |
| `Focused() bool` | Report focus state |
| `Update(tea.Msg) (tea.Model, tea.Cmd)` | Handle messages |
| `View() tea.View` | Render |
| `Value() *rune` | Current answer key, nil if unanswered |

## YesMsg / NoMsg

Emitted by a prompt built with `WithYesNo`, `WithYesNoDefaultYes`, or
`WithYesNoDefaultNo`: `YesMsg` for `'y'`/`'Y'`, `NoMsg` for `'n'`/`'N'` —
whether pressed directly or accepted via Enter as the default. Enter just
answers as if that key had been pressed; there's no separate message for it.

```go
type YesMsg struct{}
type NoMsg  struct{}
```

## Invalid keys

Pressing a key that isn't one of the accepted keys (and isn't Enter with a
default set) briefly shows that key in place of the cursor — e.g. typing `x`
against a `[y/n]` prompt shows `[y/n]: x` — without hiding the choice hint.
It clears itself automatically after `WithInvalidKeyDuration`'s duration
(600ms by default). This happens automatically in `View()`; no action is
required from the host app.

---

Sponsored by [imgproxy](https://imgproxy.net).
