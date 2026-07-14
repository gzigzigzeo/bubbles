package togglefield

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/gzigzigzeo/bubbles/form/field"
)

var keyToggle = key.NewBinding(
	key.WithKeys("space"),
	key.WithHelp("space", "toggle"),
)

// Model is a form field that represents a boolean value, rendered as a toggle switch.
type Model struct {
	field.StyledState[OnOffStyles]
	field.FocusedState
	field.DisabledState
	field.ValueState[bool]
	field.NoopInit

	width int
}

// Option configures a Model at construction time.
type Option func(*Model)

// New creates a new Model, applying opts in order.
func New(opts ...Option) *Model {
	f := &Model{}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

// SetWidth stores the field width used to pad the rendered value.
func (f *Model) SetWidth(w int) {
	f.width = w
}

// Keys returns the toggle binding.
func (f *Model) Keys() []key.Binding {
	return []key.Binding{keyToggle}
}

// Update handles key events for the Model, toggling its value when the select key is pressed.
func (f *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !f.Focused() || f.Disabled() {
		return f, nil
	}

	if k, ok := msg.(tea.KeyMsg); ok && key.Matches(k, keyToggle) {
		f.Set(!f.Get())
	}

	return f, nil
}

// View renders the Model as a tea.View showing its current state with appropriate styles.
func (f *Model) View() tea.View {
	s := f.StateStyles(f.Disabled(), f.Focused())

	if f.Get() {
		return tea.NewView(s.On.Width(f.width).Render())
	}

	return tea.NewView(s.Off.Width(f.width).Render())
}
