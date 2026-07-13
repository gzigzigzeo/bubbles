// Package buttonstack renders a fixed row of field.Controls side by side
// and manages keyboard focus among them.
package buttonstack

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/form/button"
	"github.com/gzigzigzeo/bubbles/form/field"
)

// Styles defines the visual appearance of a button row.
type Styles struct {
	Block lipgloss.Style
}

var keyLeft = key.NewBinding(
	key.WithKeys("left", "up"),
	key.WithHelp("←", "prev"),
)

var keyRight = key.NewBinding(
	key.WithKeys("right", "down"),
	key.WithHelp("→", "next"),
)

// Model is a row of Button rendered side by side. Focus/Blur/Focused are
// promoted from the embedded field.FocusState.
type Model struct {
	field.FocusState

	styles Styles
}

// New creates a Model wrapping buttons in order. It selects a button when
// the stack receives focus.
func New(buttons ...*button.Model) *Model {
	return &Model{FocusState: field.NewFocusState(buttons...)}
}

// SetStyles replaces the stack's chrome styles.
func (m *Model) SetStyles(styles Styles) {
	m.styles = styles
}

// Init calls Init on every wrapped button.
func (m *Model) Init() tea.Cmd {
	return field.Init(m.Items()...)
}

// Focus selects the first enabled button, if needed, then focuses it.
func (m *Model) Focus() tea.Cmd {
	if m.Current() == nil {
		m.FocusFirst()
	}

	return m.FocusState.Focus()
}

// Update routes left/right navigation between entries and forwards other
// messages to the focused entry.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(km, keyLeft):
			cmd, _ := m.ShiftBounded(-1)
			return m, cmd
		case key.Matches(km, keyRight):
			cmd, _ := m.ShiftBounded(1)
			return m, cmd
		}
	}

	c := m.Current()
	if c == nil {
		return m, nil
	}

	_, cmd := c.Update(msg)

	return m, cmd
}

// View renders every entry joined horizontally. Button styles control spacing.
func (m *Model) View() tea.View {
	items := m.Items()
	views := make([]string, 0, len(items))

	for _, e := range items {
		views = append(views, e.View().Content)
	}

	return tea.NewView(m.styles.Block.Render(lipgloss.JoinHorizontal(lipgloss.Top, views...)))
}

// Enable enables every entry.
func (m *Model) Enable() {
	for _, e := range m.Items() {
		e.Enable()
	}
}

// Disable disables every entry.
func (m *Model) Disable() {
	for _, e := range m.Items() {
		e.Disable()
	}
}

// Enabled reports whether any entry is enabled.
func (m *Model) Enabled() bool {
	return !m.Disabled()
}

// Disabled reports whether every entry is disabled.
func (m *Model) Disabled() bool {
	for _, e := range m.Items() {
		if !e.Disabled() {
			return false
		}
	}

	return true
}

// Keys returns left/right navigation bindings followed by the focused
// entry's own bindings.
func (m *Model) Keys() []key.Binding {
	bindings := []key.Binding{keyLeft, keyRight}

	if c := m.Current(); c != nil {
		bindings = append(bindings, c.Keys()...)
	}

	return bindings
}

// Hint returns the focused entry's hint text, if it implements field.Hinted.
func (m *Model) Hint() string {
	if h, ok := m.Current().(field.Hinted); ok {
		return h.Hint()
	}

	return ""
}
