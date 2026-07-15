// Package textinput provides a focusable, disableable text input row with a
// label. It delegates input handling to an underlying textinput.Model and
// supports an optional key filter for variants such as numeric input.
package textinput

import (
	"errors"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator/row"
)

// ErrInvalidValue is a generic validation error that WithValidator callbacks
// can return.
var ErrInvalidValue = errors.New("invalid value")

// Option configures a Model at construction time.
type Option func(*Model)

// WithFilter restricts which key messages reach the underlying input.
func WithFilter(filter func(tea.KeyMsg) bool) Option {
	return func(m *Model) {
		m.filter = filter
	}
}

// WithLabel sets the row label.
func WithLabel(label string) Option {
	return func(m *Model) {
		m.SetLabel(label)
	}
}

// WithStyles replaces the default styles.
func WithStyles(styles Styles) Option {
	return func(m *Model) {
		m.SetStyles(styles)
	}
}

// WithPlaceholder sets the placeholder text shown when the input is empty.
func WithPlaceholder(placeholder string) Option {
	return func(m *Model) {
		m.input.Placeholder = placeholder
	}
}

// WithEchoMode sets the echo mode of the underlying input (e.g.
// textinput.EchoPassword).
func WithEchoMode(mode textinput.EchoMode) Option {
	return func(m *Model) {
		m.input.EchoMode = mode
	}
}

// WithWidth sets the width of the underlying input.
func WithWidth(width int) Option {
	return func(m *Model) {
		m.input.SetWidth(width)
	}
}

// WithValidator sets a validator run against the current string value.
func WithValidator(fn func(string) error) Option {
	return func(m *Model) {
		m.validator = fn
	}
}

// Model is a focusable, disableable text input row with a label.
type Model struct {
	row.StatefulStyles[TextInputStyles]
	row.FocusedState
	row.DisabledState
	row.LabelState
	row.ErrorState

	input     textinput.Model
	filter    func(tea.KeyMsg) bool
	validator func(string) error
}

// New creates a text input row, applying opts in order.
func New(opts ...Option) *Model {
	input := textinput.New()
	input.Prompt = ""

	m := &Model{input: input}
	m.SetStyles(DefaultStyles())

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// syncInputStyles pushes the active state variant into the underlying input.
func (m *Model) syncInputStyles() {
	m.input.SetStyles(m.StateStyles(m.Disabled(), m.Focused()).Input)
}

// SetStyles replaces the current styles and pushes the active variant into the
// input.
func (m *Model) SetStyles(styles Styles) {
	m.StatefulStyles.SetStyles(styles)
	m.syncInputStyles()
}

// Enable marks the row as enabled and pushes the active variant into the input.
func (m *Model) Enable() tea.Cmd {
	m.DisabledState.Enable()
	m.syncInputStyles()

	return nil
}

// Disable marks the row as disabled and pushes the active variant into the input.
func (m *Model) Disable() tea.Cmd {
	m.DisabledState.Disable()
	m.syncInputStyles()

	return nil
}

// Focus focuses the underlying input and pushes the focused input style variant.
func (m *Model) Focus() tea.Cmd {
	m.FocusedState.Focus()
	cmd := m.input.Focus()
	m.syncInputStyles()

	return cmd
}

// Blur removes focus from the underlying input, pushes the blurred input style
// variant, and validates the current value.
func (m *Model) Blur() tea.Cmd {
	m.FocusedState.Blur()
	m.input.Blur()
	m.syncInputStyles()
	_ = m.Validate()

	return nil
}

// Focused reports whether the row is focused.
func (m *Model) Focused() bool {
	return m.FocusedState.Focused()
}

// Init satisfies tea.Model. The underlying textinput needs no asynchronous
// setup.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update forwards msg to the underlying textinput.Model, applying the optional
// key filter. Key messages are ignored while the row is disabled.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.Disabled() {
		return m, nil
	}

	if k, ok := msg.(tea.KeyMsg); ok && m.filter != nil && !m.filter(k) {
		return m, nil
	}

	updated, cmd := m.input.Update(msg)
	m.input = updated

	return m, cmd
}

// View renders the label and the underlying input horizontally joined.
func (m *Model) View() tea.View {
	styles := m.StateStyles(m.Disabled(), m.Focused())

	labelStyle := styles.Label
	if m.Err() != nil {
		labelStyle = styles.Error
	}

	label := labelStyle.Render(m.Label())
	input := m.input.View()

	line := lipgloss.JoinHorizontal(lipgloss.Top, label, input)

	return tea.NewView(line)
}

// SetWidth sets the width of the underlying input.
func (m *Model) SetWidth(width int) {
	m.input.SetWidth(width)
}

// Width returns the configured width of the underlying input.
func (m *Model) Width() int {
	return m.input.Width()
}

// EchoMode returns the echo mode of the underlying input.
func (m *Model) EchoMode() textinput.EchoMode {
	return m.input.EchoMode
}

// Validate runs the configured validator against the current value and stores
// the result. It returns nil when no validator is set or the value is valid.
func (m *Model) Validate() error {
	if m.validator == nil {
		m.SetErr(nil)

		return nil
	}

	err := m.validator(m.input.Value())
	m.SetErr(err)

	return err
}

// Get returns the current value of the input.
func (m *Model) Get() string {
	return m.input.Value()
}

// Set sets the value of the input.
func (m *Model) Set(value string) {
	m.input.SetValue(value)
}

// NumberFilter allows only numeric characters through. Navigation and editing
// keys with empty text are allowed.
func NumberFilter(k tea.KeyMsg) bool {
	text := k.Key().Text
	if text == "" {
		return true
	}

	for _, r := range text {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}
