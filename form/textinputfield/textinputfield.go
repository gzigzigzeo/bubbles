// Package textinputfield provides shared state for textinput-backed fields
// (textfield.Model, numberfield.Model).
package textinputfield

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/gzigzigzeo/bubbles/form/field"
)

// State wraps a textinput.Model with StyledState, DisabledState, NoopInit,
// and ZeroLeftPadding, shared by textfield.Model and numberfield.Model.
type State struct {
	field.StyledState[textinput.Styles]
	field.DisabledState
	field.NoopInit
	field.ZeroLeftPadding

	input  textinput.Model
	filter func(tea.KeyMsg) bool
}

// Option configures a State at construction time.
type Option func(*State)

// WithKeyFilter restricts which key messages reach the underlying input.
func WithKeyFilter(filter func(tea.KeyMsg) bool) Option {
	return func(s *State) {
		s.filter = filter
	}
}

// New creates a State with an empty prompt, applying opts in order.
func New(opts ...Option) State {
	ti := textinput.New()
	ti.Prompt = ""

	s := State{input: ti}

	for _, opt := range opts {
		opt(&s)
	}

	return s
}

// SetStyles replaces the current styles and pushes the active variant into the input.
func (s *State) SetStyles(styles Styles) {
	s.StyledState.SetStyles(styles)
	s.input.SetStyles(s.StateStyles(s.Disabled(), s.Focused()))
}

// Enable marks the field as enabled and pushes the active variant into the input.
func (s *State) Enable() {
	s.DisabledState.Enable()
	s.input.SetStyles(s.StateStyles(s.Disabled(), s.Focused()))
}

// Disable marks the field as disabled and pushes the active variant into the input.
func (s *State) Disable() {
	s.DisabledState.Disable()
	s.input.SetStyles(s.StateStyles(s.Disabled(), s.Focused()))
}

// Focus sets the underlying text input to be focused, allowing user input.
func (s *State) Focus() tea.Cmd {
	cmd := s.input.Focus()
	s.input.SetStyles(s.StateStyles(s.Disabled(), true))

	return cmd
}

// Blur removes focus from the underlying text input, preventing user input.
func (s *State) Blur() {
	s.input.Blur()
	s.input.SetStyles(s.StateStyles(s.Disabled(), false))
}

// Focused returns true if the underlying text input is currently focused.
func (s *State) Focused() bool {
	return s.input.Focused()
}

// SetWidth sets the width of the underlying text input.
func (s *State) SetWidth(w int) {
	s.input.SetWidth(w)
}

// Keys returns no field-level key bindings; text input is handled by the underlying model.
func (s *State) Keys() []key.Binding {
	return nil
}

// View renders the underlying text input as a tea.View.
func (s *State) View() tea.View {
	return tea.NewView(s.input.View())
}

// Input returns the underlying textinput.Model for direct access.
func (s *State) Input() *textinput.Model {
	return &s.input
}

// Update forwards msg to the underlying textinput.Model, filtering via the key filter.
func (s *State) Update(msg tea.Msg) tea.Cmd {
	if k, ok := msg.(tea.KeyMsg); ok && s.filter != nil && !s.filter(k) {
		return nil
	}

	m, cmd := s.input.Update(msg)
	s.input = m

	return cmd
}
