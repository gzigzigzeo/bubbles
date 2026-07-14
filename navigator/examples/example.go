// example exercises Navigator: non-focusable label rows, focusable and
// disabled item rows, a closed outer Navigator (wraps at boundaries), an open
// inner Navigator (defers boundary exit to outer), and a single external
// viewport that scrolls to nav.CursorLine().
//
// Run: go run ./navigator/examples/
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator"
	"github.com/gzigzigzeo/bubbles/scrollview"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	sectionStyle = lipgloss.NewStyle().Faint(true)
	hintStyle    = lipgloss.NewStyle().Faint(true).MarginTop(1)
	cursorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)

// ─── Row types ───────────────────────────────────────────────────────────────

// label is a non-focusable, non-disableable display row.
type label string

// Init satisfies tea.Model.
func (l label) Init() tea.Cmd {
	return nil
}

// Update satisfies tea.Model.
func (l label) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return l, nil
}

// View renders the label text with faint styling.
func (l label) View() tea.View {
	return tea.NewView(sectionStyle.Render(" " + string(l)))
}

// item is a toggleable checkbox row; implements Focusable and Disableable.
type item struct {
	text     string
	checked  bool
	focused  bool
	disabled bool
	indent   string
}

// Init satisfies tea.Model.
func (it *item) Init() tea.Cmd {
	return nil
}

// Update toggles the checkbox when space is pressed while focused.
func (it *item) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok || !it.focused {
		return it, nil
	}

	if km.String() == "space" {
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
	nav       *navigator.Model
	vp        scrollview.Model
	maxHeight int
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
		&item{text: "Echo", indent: "  "},
		&item{text: "Foxtrot", indent: "  "},
		&item{text: "Golf", indent: "  "},
		&item{text: "Hotel", indent: "  "},
		&item{text: "India", indent: "  "},
	)

	outer := navigator.New(
		label("─ Items (outer navigator, closed) ─"),
		&item{text: "Alpha"},
		&item{text: "Beta"},
		&item{
			text:     "Gamma  (disabled — skipped by focus)",
			disabled: true,
		},
		&item{text: "Delta"},
		label("─ Inner navigator (open) ─"),
		inner,
	)

	vp := scrollview.New()
	vp.SetWidth(43)

	outer.ViewportCoordinator().SetHeight(7)
	outer.ViewportCoordinator().SetViewport(&vp)

	return &model{
		nav:       outer,
		vp:        vp,
		maxHeight: 7,
	}
}

// Init focuses the first row and initialises all rows.
func (m *model) Init() tea.Cmd {
	return tea.Batch(m.nav.FocusFirst(), m.nav.Init(), m.vp.Init())
}

// Update routes messages to the navigator.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok && km.String() == "q" {
		return m, tea.Quit
	}

	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.maxHeight = max(1, ws.Height-4)
		m.vp.SetWidth(ws.Width)
		m.syncViewportHeight()

		return m, nil
	}

	updated, cmd := m.nav.Update(msg)
	m.nav = updated.(*navigator.Model)
	m.syncViewportHeight()

	return m, cmd
}

// syncViewportHeight sets the viewport and coordinator heights to the smaller
// of the available screen space and the navigator's current content height.
func (m *model) syncViewportHeight() {
	contentHeight := m.nav.Height()
	h := max(1, min(m.maxHeight, contentHeight))
	m.nav.ViewportCoordinator().SetHeight(h)
}

// View syncs the navigator's flat output into the viewport and composes the
// final screen.
func (m *model) View() tea.View {
	m.nav.ViewportCoordinator().SetContent(m.nav.View().Content)

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Navigator Demo"),
		m.nav.ViewportCoordinator().View().Content,
		"",
		cursorStyle.Render(fmt.Sprintf("cursor line: %d", m.nav.CursorLine())),
		hintStyle.Render("↑/↓/k/j: navigate   space: toggle   q: quit"),
	)

	return tea.NewView(content)
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
