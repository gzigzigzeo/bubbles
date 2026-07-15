# Code Style

## Methods

- Every exported and unexported method must have a Go doc comment (`// MethodName ...`).
- Never write single-line method bodies. Always use a multi-line block:
  ```go
  // Bad
  func (f *Foo) Bar() string { return f.bar }

  // Good
  func (f *Foo) Bar() string {
      return f.bar
  }
  ```

## Struct Literals

- When a struct literal has more than one field, put each field on its own line:
  ```go
  // Bad
  Entry{Label: "Port", Hint: "...", Field: port}

  // Good
  Entry{
      Label: "Port",
      Hint:  "...",
      Field: port,
  }
  ```

## Functions

Disallow one-liner syntax
  ```go
  // Bad
  func (l Label) Text() string { return "" }

  // Good
  func (l Label) Text() string {
    return ""
  }
  ```

## Do not use tmux for debug

Do not run any kind of visual debugging.

## Add empty line } after block

Add empty line after } block is closed (if, for exmaple) and before return statement.

```go
// Bad
if x {
    ...
}
return y

// Good
if x {
    ...
}

return y
```

## Add empty line before if

Add empty line before if statement if there is none.

```go
// Bad
x := 1
if x > 0 {
    ...
}

// Good
x := 1

if x > 0 {
    ...
}
```

## Comment

Limit comments to 2-3 lines max. Do not comment test methods, they should be self-explanatory.

## Conditions

Prefer guard clauses.

```go
// Bad
for range i {
  if i == 5 {
    n := i * 2
    n = math.Pow(n, 5)
    return x
  }
}

// Good
for range i {
  if i != 5 {
    continue
  }

n := i * 2
  n = math.Pow(n, 5)
  return x
}
```

## Keybindings

Use `charm.land/bubbles/v2/key` bindings and `key.Matches` instead of comparing `tea.KeyMsg.String()`.

Declare default key bindings as package-level variables, not functions or methods. Both `key.WithKeys` and `key.WithHelp` are required: `WithHelp` only sets help text and does not define the keys that `key.Matches` checks.

```go
// Bad
switch km.String() {
case "up", "k":
    // ...
}

// Bad
func keyUpBinding() key.Binding {
    return key.NewBinding(
        key.WithKeys("up", "k"),
        key.WithHelp("↑/k", "up"),
    )
}

// Good
var navKeyUp = key.NewBinding(
    key.WithKeys("up", "k"),
    key.WithHelp("↑/k", "up"),
)

switch {
case key.Matches(km, navKeyUp):
    // ...
}
```

## Do not use \n and spaces in views

Instead of `strings.Join(rows, "\n")` use `lipgloss.JoinVertical`
Instead of strings.Repeat(" ", 5) define it as a style with Width.

```go
// Bad
return tea.NewView(strings.Join(rows, "\n"))

// Good
return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, rows...))
```

```go
// Bad
return tea.NewView(strings.Repeat(" ", 5) + text)

// Good
style := lipgloss.NewStyle().Width(5) // Prefer existing styles

return tea.NewView(style.Render(text))
```

## Don't use t.Fatalf in tests

Instead, use require.*

```go
// Bad
if got != want {
    t.Fatalf("expected %d, got %d", want, got)
}

// Good
require.Equal(t, want, got)
```

## Do not provide extra descriptions for require.* in tests

Do not do that: require.Equal(x, 3, "description")

```go
// Bad
require.Equal(t, want, got, "values should match")

// Good
require.Equal(t, want, got)
```

## Examples

Every example must live in its own isolated Go module with its own `go.mod`. The directory must be named `example`. Examples reference packages in this repository by local path using `replace` directives, not by published version.

## Row / Collection separation

In navigator-based components, keep rows as data sources and collections as behaviour owners.

- **Row** — holds data, view state, focus/disabled state, and the message to emit when activated. It does not handle activation keys itself.
- **Collection** — owns activation semantics: key bindings, selection mode, focused-index tracking, and message emission. Rows transparently forward keys to their collection.

```go
// Good: row is data; collection handles Enter/Space.
row := menu.New("Alpha", "alpha", "", MySelectMsg{Value: "alpha"})
collection := menu.NewController(rows, menu.WithMode[string](menu.ModeMultiSelect))
```

Active interactive rows such as dropdowns or text inputs may consume their own keys. The navigator itself only moves focus and forwards unhandled keys; it does not know about collections or activation semantics.
