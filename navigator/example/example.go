// example exercises Navigator: non-focusable label rows, focusable and
// disabled item rows, a closed outer Navigator (wraps at boundaries), an open
// inner Navigator (defers boundary exit to outer), and a single external
// viewport that scrolls to nav.CursorLine().
//
// Run: go run ./navigator/example/
package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator"
	"github.com/gzigzigzeo/bubbles/scrollview"
)

const (
	defaultWidth  = 43
	defaultHeight = 7
	heightPadding = 4
)

// ─── Row types ───────────────────────────────────────────────────────────────

// label is a non-focusable, non-disableable display row.
type label struct {
	text  string
	style lipgloss.Style
}

// newLabel creates a label row rendered with the given style.
func newLabel(text string, style lipgloss.Style) label {
	return label{
		text:  text,
		style: style,
	}
}

// Init satisfies tea.Model.
func (l label) Init() tea.Cmd {
	return nil
}

// Update satisfies tea.Model.
func (l label) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return l, nil
}

// View renders the label text with the configured style.
func (l label) View() tea.View {
	return tea.NewView(l.style.Render(" " + l.text))
}

// item is a toggleable checkbox row; implements Focusable and Disableable.
type item struct {
	text     string
	checked  bool
	focused  bool
	disabled bool
	indent   string
}

// newItem creates a top-level, enabled item row.
func newItem(text string) *item {
	return &item{
		text:     text,
		checked:  false,
		focused:  false,
		disabled: false,
		indent:   "",
	}
}

// newInnerItem creates an enabled item row indented for the inner navigator.
func newInnerItem(text string) *item {
	return &item{
		text:     text,
		checked:  false,
		focused:  false,
		disabled: false,
		indent:   "  ",
	}
}

// newDisabledItem creates a disabled item row.
func newDisabledItem(text string) *item {
	return &item{
		text:     text,
		checked:  false,
		focused:  false,
		disabled: true,
		indent:   "",
	}
}

// Init satisfies tea.Model.
func (it *item) Init() tea.Cmd {
	return nil
}

// Update toggles the checkbox when space is pressed while focused.
func (it *item) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok || !it.focused {
		return it, nil
	}

	toggleKey := key.NewBinding(
		key.WithKeys("space"),
		key.WithHelp("space", "toggle"),
	)
	if key.Matches(keyMsg, toggleKey) {
		it.checked = !it.checked
	}

	return it, nil
}

// View renders the cursor indicator, checkbox, and text.
func (it *item) View() tea.View {
	cursor := "  "
	if it.focused {
		cursor = "▶ "
	}

	check := "[ ]"
	if it.checked {
		check = "[✓]"
	}

	line := it.indent + " " + cursor + check + " " + it.text

	if it.disabled {
		line = lipgloss.NewStyle().Faint(true).Render(line)
	}

	return tea.NewView(line)
}

// Focus marks the item as focused.
func (it *item) Focus() tea.Cmd {
	it.focused = true

	return nil
}

// Blur removes focus.
func (it *item) Blur() tea.Cmd {
	it.focused = false

	return nil
}

// Focused reports focus state.
func (it *item) Focused() bool {
	return it.focused
}

// Enable marks the item as enabled.
func (it *item) Enable() tea.Cmd {
	it.disabled = false

	return nil
}

// Disable marks the item as disabled.
func (it *item) Disable() tea.Cmd {
	it.disabled = true

	return nil
}

// Disabled reports whether the item is disabled.
func (it *item) Disabled() bool {
	return it.disabled
}

// ─── Root model ──────────────────────────────────────────────────────────────

// model is the root Bubble Tea model.
type model struct {
	nav         *navigator.Model
	viewport    scrollview.Model
	maxHeight   int
	titleStyle  lipgloss.Style
	hintStyle   lipgloss.Style
	cursorStyle lipgloss.Style
}

// newModel builds the demo.
//
// Outer navigator — open (focus stays at boundaries, viewport scrolls):
//   - label row (non-focusable heading)
//   - three item rows (Alpha, Beta, Delta)
//   - one disabled item (Gamma — skipped by focus)
//   - label row
//   - inner open navigator (five items — defers boundary scroll to outer)
//
// A single viewport (height=7) is attached to the navigator's internal
// ViewportCoordinator. The coordinator keeps the viewport offset in sync on
// every update.
func newModel() *model {
	inner := navigator.New(
		newInnerItem("Echo"),
		newInnerItem("Foxtrot"),
		newInnerItem("Golf"),
		newInnerItem("Hotel"),
		newInnerItem("India"),
	)

	sectionStyle := lipgloss.NewStyle().Faint(true)

	outer := navigator.New(
		newLabel("─ Items (outer navigator, closed) ─", sectionStyle),
		newItem("Alpha"),
		newItem("Beta"),
		newDisabledItem("Gamma  (disabled — skipped by focus)"),
		newItem("Delta"),
		newLabel("─ Inner navigator (open) ─", sectionStyle),
		inner,
	)

	viewport := scrollview.New()
	viewport.SetWidth(defaultWidth)

	outer.ViewportCoordinator().SetHeight(defaultHeight)
	outer.ViewportCoordinator().SetViewport(&viewport)

	return &model{
		nav:         outer,
		viewport:    viewport,
		maxHeight:   defaultHeight,
		titleStyle:  lipgloss.NewStyle().Bold(true).MarginBottom(1),
		hintStyle:   lipgloss.NewStyle().Faint(true).MarginTop(1),
		cursorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
	}
}

// Init focuses the first row and initialises all rows.
func (m *model) Init() tea.Cmd {
	return tea.Batch(m.nav.FocusFirst(), m.nav.Init(), m.viewport.Init())
}

// Update routes messages to the navigator.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		quitKey := key.NewBinding(key.WithKeys("q"))
		if key.Matches(km, quitKey) {
			return m, tea.Quit
		}
	}

	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.maxHeight = max(1, ws.Height-heightPadding)
		m.viewport.SetWidth(ws.Width)
		m.syncViewportHeight()

		return m, nil
	}

	updated, cmd := m.nav.Update(msg)

	updatedNav, ok := updated.(*navigator.Model)
	if !ok {
		return m, cmd
	}

	m.nav = updatedNav
	m.syncViewportHeight()

	return m, cmd
}

// View composes the final screen from the navigator's clipped viewport output.
func (m *model) View() tea.View {
	content := lipgloss.JoinVertical(lipgloss.Left,
		m.titleStyle.Render("Navigator Demo"),
		m.nav.ViewportCoordinator().View(),
		"",
		m.cursorStyle.Render(fmt.Sprintf("cursor line: %d", m.nav.CursorLine())),
		m.hintStyle.Render("↑/↓/k/j: navigate   space: toggle   q: quit"),
	)

	return tea.NewView(content)
}

// syncViewportHeight sets the viewport and coordinator heights to the smaller
// of the available screen space and the navigator's current content height.
func (m *model) syncViewportHeight() {
	contentHeight := m.nav.Height()
	h := max(1, min(m.maxHeight, contentHeight))
	m.nav.ViewportCoordinator().SetHeight(h)
}

func main() {
	_, err := tea.NewProgram(newModel()).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
