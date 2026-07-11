// example demonstrates a 10-item menu with only 5 rows visible at once, a
// pre-set marker showing a separately committed value, and the cursor
// scrolling past that window.
// Run: go run ./examples
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/menu"
)

const menuWidth = 40

var fruits = []string{
	"Apple", "Banana", "Cherry", "Date", "Elderberry",
	"Fig", "Grape", "Honeydew", "Kiwi", "Lemon",
}

type model struct {
	menu    *menu.Menu[string]
	chosen  string
	styleHd lipgloss.Style
}

func newModel() model {
	opts := make([]menu.Option[string], len(fruits))
	for i, name := range fruits {
		opts[i] = menu.Option[string]{Name: name, Value: name}
	}

	m := menu.New(opts)
	m.SetStyles(menu.DefaultStyles())
	// SetWidth must be called explicitly to claim the widest available
	// space; New defaults to a fixed 80 columns otherwise.
	m.SetWidth(menuWidth)
	// Only the first 5 of the 10 items are visible at once; the rest
	// scroll into view as the cursor moves.
	m.SetHeight(5)
	// Show "Cherry" as the separately committed value while the cursor
	// starts elsewhere, to demonstrate the marker.
	m.SetMarker("Cherry")

	return model{
		menu:    m,
		styleHd: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#53d1ff")).MarginBottom(1),
	}
}

func (m model) Init() tea.Cmd {
	return m.menu.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok && km.String() == "q" {
		return m, tea.Quit
	}

	if choice, ok := msg.(menu.ChoiceMsg[string]); ok {
		m.chosen = choice.Value
		return m, nil
	}

	_, cmd := m.menu.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	header := m.styleHd.Render("    menu — 10 items, 5 visible")

	content := header + "\n" + m.menu.View().Content

	if m.chosen != "" {
		content += "\n\n    Selected: " + m.chosen
	}

	return tea.NewView(content)
}

func main() {
	prog := tea.NewProgram(newModel())
	if _, err := prog.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
