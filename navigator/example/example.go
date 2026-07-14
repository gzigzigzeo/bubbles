// example demonstrates Navigator with button rows and a horizontal button stack.
// Navigate with ↑/↓ or k/j, press a row with enter, and move between
// buttons with ←/→ or h/l and press with enter.
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
	"github.com/gzigzigzeo/bubbles/navigator/rows/button"
	"github.com/gzigzigzeo/bubbles/scrollview"
)

const (
	defaultWidth  = 43
	defaultHeight = 7
	heightPadding = 4
)

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

// selectMsg is emitted when a menu row is selected.
type selectMsg struct {
	Value string
}

// pressMsg is emitted when a button is pressed.
type pressMsg struct {
	Label string
}

// model is the root Bubble Tea model.
type model struct {
	nav         *navigator.Model
	viewport    scrollview.Model
	maxHeight   int
	titleStyle  lipgloss.Style
	hintStyle   lipgloss.Style
	cursorStyle lipgloss.Style
	statusStyle lipgloss.Style
	status      string
}

// newModel builds the demo.
//
// Outer navigator — open (focus stays at boundaries, viewport scrolls):
//   - heading label
//   - first group of selectable menu rows
//   - heading label
//   - second group of selectable menu rows
//   - horizontal button stack with three buttons
//
// A single viewport (height=7) is attached to the navigator's internal
// ViewportController. The controller keeps the viewport offset in sync on
// every update.
func newModel() *model {
	sectionStyle := lipgloss.NewStyle().Faint(true)

	firstGroup := []*menurow.Model[string]{
		menurow.New("Alpha", "alpha", "First option", selectMsg{Value: "alpha"}),
		menurow.New("Beta", "beta", "Second option", selectMsg{Value: "beta"}),
		menurow.New("Gamma", "gamma", "Third option", selectMsg{Value: "gamma"}),
	}
	_ = menurow.NewController(firstGroup, menurow.WithMode[string](menurow.ModeSelect))

	secondGroup := []*menurow.Model[string]{
		menurow.New("Delta", "delta", "Fourth option", selectMsg{Value: "delta"}),
		menurow.New("Epsilon", "epsilon", "Fifth option", selectMsg{Value: "epsilon"}),
	}
	_ = menurow.NewController(secondGroup, menurow.WithMode[string](menurow.ModeSelect))

	buttonStack := button.NewStack(
		button.New("Save", pressMsg{Label: "Save"}),
		button.New("Cancel", pressMsg{Label: "Cancel"}),
		button.New("Help", pressMsg{Label: "Help"}),
	)
	buttonStack.SetStyles(button.StackStyles{
		Wrapper: lipgloss.NewStyle().MarginTop(1),
	})

	teaRows := make([]tea.Model, 0,
		1+len(firstGroup)+1+len(secondGroup)+1,
	)
	teaRows = append(teaRows, newLabel("─ Group 1 ─", sectionStyle))
	for _, r := range firstGroup {
		teaRows = append(teaRows, r)
	}
	teaRows = append(teaRows, newLabel("─ Group 2 ─", sectionStyle))
	for _, r := range secondGroup {
		teaRows = append(teaRows, r)
	}
	teaRows = append(teaRows, buttonStack)

	outer := navigator.New(teaRows...)

	viewport := scrollview.New()
	viewport.SetWidth(defaultWidth)

	outer.ViewportController().SetHeight(defaultHeight)
	outer.ViewportController().SetViewport(&viewport)

	return &model{
		nav:         outer,
		viewport:    viewport,
		maxHeight:   defaultHeight,
		titleStyle:  lipgloss.NewStyle().Bold(true).MarginBottom(1),
		hintStyle:   lipgloss.NewStyle().Faint(true).MarginTop(1),
		cursorStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		statusStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#53d1ff")),
	}
}

// Init focuses the first row and initialises all rows.
func (m *model) Init() tea.Cmd {
	return tea.Batch(m.nav.FocusFirst(), m.nav.Init(), m.viewport.Init())
}

// Update routes messages to the navigator.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case selectMsg:
		m.status = fmt.Sprintf("Selected: %s", msg.Value)

		return m, nil

	case pressMsg:
		m.status = fmt.Sprintf("Pressed: %s", msg.Label)

		return m, nil

	case tea.KeyMsg:
		quitKey := key.NewBinding(key.WithKeys("q"))
		if key.Matches(msg, quitKey) {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.maxHeight = max(1, msg.Height-heightPadding)
		m.viewport.SetWidth(msg.Width)
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
		m.nav.ViewportController().View(),
		"",
		m.cursorStyle.Render(fmt.Sprintf("cursor line: %d", m.nav.CursorLine())),
		m.hintStyle.Render("↑/↓/k/j: navigate   enter: select   ←/→/h/l: buttons   q: quit"),
	)

	if m.status != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			m.statusStyle.Render(m.status),
		)
	}

	return tea.NewView(content)
}

// syncViewportHeight sets the viewport and controller heights to the smaller
// of the available screen space and the navigator's current content height.
func (m *model) syncViewportHeight() {
	contentHeight := m.nav.Height()
	h := max(1, min(m.maxHeight, contentHeight))
	m.nav.ViewportController().SetHeight(h)
}

func main() {
	_, err := tea.NewProgram(newModel()).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
