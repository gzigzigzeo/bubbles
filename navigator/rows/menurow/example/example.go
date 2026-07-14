// example demonstrates menurow rows inside a navigator: navigate with ↑/↓ or
// k/j, mark/unmark rows with space, and select a row with enter.
//
// Run: go run ./navigator/rows/menurow/example/
package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator"
	"github.com/gzigzigzeo/bubbles/navigator/rows/menurow"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	hintStyle   = lipgloss.NewStyle().Faint(true).MarginTop(1)
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#53d1ff"))
)

var quitKey = key.NewBinding(
	key.WithKeys("q"),
)

// selectMsg is the application-specific message emitted when a row is selected.
type selectMsg struct {
	Value string
}

// model is the example root. It owns the navigator and the menu row controller.
type model struct {
	nav        *navigator.Model
	controller *menurow.Controller[string]
	status     string
}

// newModel builds a navigator over a few menu rows and a controller to manage
// their marks.
func newModel() *model {
	rows := []*menurow.Model[string]{
		menurow.New("Alpha", "alpha", "First option", selectMsg{Value: "alpha"}),
		menurow.New("Beta", "beta", "Second option", selectMsg{Value: "beta"}),
		menurow.New("Gamma", "gamma", "Third option", selectMsg{Value: "gamma"}),
		menurow.New("Delta", "delta", "", selectMsg{Value: "delta"}),
		menurow.New("Epsilon", "epsilon", "Last option", selectMsg{Value: "epsilon"}),
	}

	controller := menurow.NewController(rows, menurow.WithMode[string](menurow.ModeMultiSelect))

	teaRows := make([]tea.Model, len(rows))
	for i, r := range rows {
		teaRows[i] = r
	}

	return &model{
		nav:        navigator.New(teaRows...),
		controller: controller,
	}
}

// Init focuses the first row and initializes the navigator.
func (m *model) Init() tea.Cmd {
	return tea.Batch(m.nav.FocusFirst(), m.nav.Init())
}

// Update handles quit, selection messages, and navigator navigation.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case selectMsg:
		m.status = fmt.Sprintf("Selected: %s", msg.Value)

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
		titleStyle.Render("Menu Row Demo"),
		m.nav.View().Content,
		hintStyle.Render("↑/↓/k/j: navigate   space: mark   enter: select   q: quit"),
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
