// Package button provides a focusable, disableable button row and a horizontal
// stack of buttons. A single button emits a configurable message when pressed.
package button

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator/row"
)

// defaultPressKey is the default key binding that presses a button.
var defaultPressKey = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "press"),
)

// Option configures a [Model].
type Option func(*Model)

// WithPressKeys sets the key binding that presses the button.
func WithPressKeys(keys ...string) Option {
	return func(m *Model) {
		m.pressKey = key.NewBinding(key.WithKeys(keys...), key.WithHelp(keys[0], "press"))
	}
}

// Model is a single focusable, disableable button row. It emits the configured
// message when the press key is activated while the button is focused and
// enabled.
type Model struct {
	row.StatefulStyles[lipgloss.Style]
	row.FocusedState
	row.DisabledState
	row.LabelState

	msg      tea.Msg
	pressKey key.Binding
}

// New creates a button with the given label and message. The default press key
// is enter. msg must not be nil.
func New(label string, msg tea.Msg, opts ...Option) *Model {
	if msg == nil {
		panic("button.New: msg must not be nil")
	}

	buttonModel := &Model{
		StatefulStyles: row.StatefulStyles[lipgloss.Style]{},
		FocusedState:   row.FocusedState{},
		DisabledState:  row.DisabledState{},
		LabelState:     row.LabelState{},
		msg:            msg,
		pressKey:       defaultPressKey,
	}
	buttonModel.SetLabel(label)
	buttonModel.SetStyles(DefaultStyles())

	for _, opt := range opts {
		opt(buttonModel)
	}

	return buttonModel
}

// Init satisfies [tea.Model].
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update emits the configured message when the press key is activated while the
// button is focused and enabled.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.Disabled() || !m.Focused() {
		return m, nil
	}

	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if key.Matches(km, m.pressKey) {
		return m, func() tea.Msg {
			return m.msg
		}
	}

	return m, nil
}

// View renders the button label with the state-appropriate style.
func (m *Model) View() tea.View {
	style := m.StateStyles(m.Disabled(), m.Focused())

	return tea.NewView(style.Render(m.Label()))
}

// Msg returns the message emitted when the button is pressed.
func (m *Model) Msg() tea.Msg {
	return m.msg
}

// SetMsg sets the message emitted when the button is pressed. msg must not be
// nil.
func (m *Model) SetMsg(msg tea.Msg) {
	if msg == nil {
		panic("button.Model.SetMsg: msg must not be nil")
	}

	m.msg = msg
}
