package menurow

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator/rows/row"
)

// Model is a single selectable menu row: a name, an optional description, a
// value, and the message to emit when the row is selected. It implements
// [tea.Model], [row.Focusable], and [row.Disableable].
//
// A row is intended to be owned by a [Controller]. The controller handles
// activation keys and emits the row's configured message.
type Model[T comparable] struct {
	row.StatefulStyles[RowStyles]
	row.FocusedState
	row.DisabledState

	name        string
	value       T
	description string
	marked      bool
	msg         tea.Msg // message emitted when the row is selected

	controller *Controller[T]
	index      int
}

// New creates a menu row. The msg is emitted by the owning [Controller] when
// the row is focused and the user activates it. msg must not be nil.
func New[T comparable](name string, value T, description string, msg tea.Msg) *Model[T] {
	if msg == nil {
		panic("menurow.New: msg must not be nil")
	}

	m := &Model[T]{
		name:        name,
		value:       value,
		description: description,
		msg:         msg,
	}
	m.SetStyles(DefaultStyles())

	return m
}

// Init satisfies [tea.Model].
func (m *Model[T]) Init() tea.Cmd {
	return nil
}

// Update forwards key messages to the owning [Controller]. All other messages
// are ignored.
func (m *Model[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.controller == nil || m.Disabled() {
		return m, nil
	}

	cmd := m.controller.Update(msg)

	return m, cmd
}

// View renders the row's cursor indicator, mark, name, and description.
func (m *Model[T]) View() tea.View {
	styles := m.StateStyles(m.Disabled(), m.Focused())

	cursor := styles.Cursor.Render()
	mark := m.markRender(styles)
	name := styles.Name.Render(m.name)

	parts := []string{cursor, mark, name}

	if m.description != "" {
		desc := styles.Description.Render(m.description)
		parts = append(parts, desc)
	}

	line := lipgloss.JoinHorizontal(lipgloss.Top, parts...)

	return tea.NewView(line)
}

// markRender returns the mark glyph when the row is marked, or a same-width
// blank placeholder otherwise.
func (m *Model[T]) markRender(styles RowStyles) string {
	if m.marked {
		return styles.Mark.Render()
	}

	return lipgloss.NewStyle().Width(lipgloss.Width(styles.Mark.Render())).Render("")
}

// Value returns the row's value.
func (m *Model[T]) Value() T {
	return m.value
}

// SetMarked sets whether the row is marked.
func (m *Model[T]) SetMarked(marked bool) {
	m.marked = marked
}

// Marked reports whether the row is marked.
func (m *Model[T]) Marked() bool {
	return m.marked
}

// Focus marks the row as focused and notifies the owning [Controller].
func (m *Model[T]) Focus() tea.Cmd {
	m.FocusedState.Focus()

	if m.controller != nil {
		m.controller.setFocused(m.index)
	}

	return nil
}

// Blur removes focus from the row and notifies the owning [Controller].
func (m *Model[T]) Blur() tea.Cmd {
	m.FocusedState.Blur()

	if m.controller != nil {
		m.controller.clearFocus(m.index)
	}

	return nil
}
