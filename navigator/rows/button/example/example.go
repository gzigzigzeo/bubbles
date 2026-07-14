// example demonstrates button rows and a horizontal button stack inside a
// navigator. Navigate with ↑/↓ or k/j, move between buttons with ←/→ or h/l,
// and press a button with enter.
//
// Run: go run ./navigator/rows/button/example/
package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator"
	"github.com/gzigzigzeo/bubbles/navigator/rows/button"
)

var quitKey = key.NewBinding(
	key.WithKeys("q"),
)

// pressMsg is emitted when a button is pressed.
type pressMsg struct {
	Label string
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	hintStyle   = lipgloss.NewStyle().Faint(true).MarginTop(1)
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#53d1ff"))
)

// model is the example root.
type model struct {
	nav    *navigator.Model
	status string
}

// newModel builds a navigator with a few regular rows and a button stack.
func newModel() *model {
	stack := button.NewStack(
		button.New("OK", pressMsg{Label: "OK"}),
		button.New("Cancel", pressMsg{Label: "Cancel"}),
		button.New("Help", pressMsg{Label: "Help"}),
	)

	nav := navigator.New(
		newLabelRow("Button Row Demo"),
		newLabelRow("Static row 1"),
		newLabelRow("Static row 2"),
		newLabelRow("Use ↑/↓ to navigate rows"),
		newLabelRow("Use ←/→ inside the stack"),
		stack,
	)

	return &model{
		nav: nav,
	}
}

// labelRow is a simple non-focusable display row.
type labelRow struct {
	text string
}

func newLabelRow(text string) labelRow {
	return labelRow{text: text}
}

func (l labelRow) Init() tea.Cmd {
	return nil
}

func (l labelRow) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return l, nil
}

func (l labelRow) View() tea.View {
	return tea.NewView(l.text)
}

// Init focuses the first row and initializes the navigator.
func (m *model) Init() tea.Cmd {
	return tea.Batch(m.nav.FocusFirst(), m.nav.Init())
}

// Update handles quit, button press messages, and navigator navigation.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case pressMsg:
		m.status = fmt.Sprintf("Pressed: %s", msg.Label)

		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, quitKey) {
			return m, tea.Quit
		}
	}

	updated, cmd := m.nav.Update(msg)
	if nav, ok := updated.(*navigator.Model); ok {
		m.nav = nav
	}

	return m, cmd
}

// View renders the title, rows, hint, and status line.
func (m *model) View() tea.View {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Button Demo"),
		m.nav.View().Content,
		hintStyle.Render("↑/↓/k/j: navigate   ←/→/h/l: buttons   enter: press   q: quit"),
	)

	if m.status != "" {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			statusStyle.Render(m.status),
		)
	}

	return tea.NewView(content)
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
