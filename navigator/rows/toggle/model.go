package toggle

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator/row"
)

// OnMsg is emitted when the toggle becomes true.
type OnMsg struct{}

// OffMsg is emitted when the toggle becomes false.
type OffMsg struct{}

// defaultToggleKey is the default key binding that toggles the row.
var defaultToggleKey = key.NewBinding(
	key.WithKeys("space"),
	key.WithHelp("space", "toggle"),
)

// Option configures a Model at construction time.
type Option func(*Model)

// WithToggleKeys sets the key binding that toggles the row.
func WithToggleKeys(keys ...string) Option {
	return func(m *Model) {
		m.toggleKey = key.NewBinding(key.WithKeys(keys...), key.WithHelp(keys[0], "toggle"))
	}
}

// WithStyles replaces the default styles.
func WithStyles(styles Styles) Option {
	return func(m *Model) {
		m.SetStyles(styles)
	}
}

// WithValue sets the initial value.
func WithValue(value bool) Option {
	return func(m *Model) {
		m.value = value
	}
}

// Model is a focusable, disableable toggle row with a label.
type Model struct {
	row.StatefulStyles[ToggleStyles]
	row.FocusedState
	row.DisabledState
	row.LabelState

	value     bool
	toggleKey key.Binding
}

// New creates a toggle row with the given label.
func New(label string, opts ...Option) *Model {
	m := &Model{
		StatefulStyles: row.StatefulStyles[ToggleStyles]{},
		FocusedState:   row.FocusedState{},
		DisabledState:  row.DisabledState{},
		LabelState:     row.LabelState{},
		toggleKey:      defaultToggleKey,
	}
	m.SetLabel(label)
	m.SetStyles(DefaultStyles("On", "Off"))

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// Init satisfies tea.Model.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update toggles the value when the toggle key is pressed while the row is
// focused and enabled. When the value changes, it returns a command that emits
// either [OnMsg] or [OffMsg].
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.Focused() || m.Disabled() {
		return m, nil
	}

	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if key.Matches(km, m.toggleKey) {
		m.value = !m.value

		if m.value {
			return m, func() tea.Msg {
				return OnMsg{}
			}
		}

		return m, func() tea.Msg {
			return OffMsg{}
		}
	}

	return m, nil
}

// View renders the label and the current on/off state.
func (m *Model) View() tea.View {
	styles := m.StateStyles(m.Disabled(), m.Focused())

	label := styles.Label.Render(m.Label())

	var toggle string
	if m.value {
		toggle = styles.On.Render()
	} else {
		toggle = styles.Off.Render()
	}

	line := lipgloss.JoinHorizontal(lipgloss.Top, label, toggle)

	return tea.NewView(line)
}

// Keys returns the toggle binding.
func (m *Model) Keys() []key.Binding {
	return []key.Binding{m.toggleKey}
}

// Get returns the current value.
func (m *Model) Get() bool {
	return m.value
}

// Set sets the current value.
func (m *Model) Set(value bool) {
	m.value = value
}
