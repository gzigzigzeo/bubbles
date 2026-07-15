// Package stack provides a simple layout component that arranges [tea.Model]
// items vertically or horizontally. It forwards every message to all of its
// items and combines their commands into a [tea.Sequence].
package stack

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Mode determines how items are laid out.
type Mode int

const (
	// Vertical arranges items from top to bottom.
	Vertical Mode = iota
	// Horizontal arranges items from left to right.
	Horizontal
)

// crossAxis positions used when joining item views.
const (
	verticalCrossAxis   = lipgloss.Left
	horizontalCrossAxis = lipgloss.Top
)

// Model lays out a collection of [tea.Model] items in a single direction.
type Model struct {
	items []tea.Model
	mode  Mode
}

// NewVertical creates a vertically arranged stack.
func NewVertical(items ...tea.Model) *Model {
	return &Model{
		items: items,
		mode:  Vertical,
	}
}

// NewHorizontal creates a horizontally arranged stack.
func NewHorizontal(items ...tea.Model) *Model {
	return &Model{
		items: items,
		mode:  Horizontal,
	}
}

// Items returns the stack's items.
func (m *Model) Items() []tea.Model {
	return m.items
}

// Init initializes every item and returns the combined commands as a sequence.
func (m *Model) Init() tea.Cmd {
	_, cmd := ForEach(m.items, func(it tea.Model) (tea.Model, tea.Cmd) {
		return it, it.Init()
	})

	return cmd
}

// Update forwards the message to every item, replacing each with its updated
// model, and returns their commands as a sequence.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	items, cmd := ForEach(m.items, func(it tea.Model) (tea.Model, tea.Cmd) {
		return it.Update(msg)
	})
	m.items = items

	return m, cmd
}

// View renders the items joined vertically or horizontally.
func (m *Model) View() tea.View {
	parts := make([]string, 0, len(m.items))

	for _, it := range m.items {
		if it == nil {
			continue
		}

		parts = append(parts, it.View().Content)
	}

	switch m.mode {
	case Horizontal:
		return tea.NewView(lipgloss.JoinHorizontal(horizontalCrossAxis, parts...))
	default:
		return tea.NewView(lipgloss.JoinVertical(verticalCrossAxis, parts...))
	}
}
