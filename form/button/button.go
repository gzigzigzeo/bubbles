package button

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/form/field"
)

var keyActivate = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "activate"),
)

// Model is a focusable, disableable widget that emits a command when activated.
type Model struct {
	field.FocusedState
	field.DisabledState
	field.NoopInit
	field.StyledState[lipgloss.Style]

	label string
	cmd   tea.Cmd

	// HintText is shown while the button is focused.
	HintText string
}

// Option configures a Model at construction time.
type Option func(*Model)

// New creates a Model with the given label. When activated, cmd is returned.
func New(label string, cmd tea.Cmd, opts ...Option) *Model {
	b := &Model{label: label, cmd: cmd}

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// Init returns nil.
func (b *Model) Init() tea.Cmd {
	return nil
}

// Keys returns the activation binding.
func (b *Model) Keys() []key.Binding {
	return []key.Binding{keyActivate}
}

// Update processes key messages and returns the button's command when activated.
func (b *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !b.Focused() || b.Disabled() {
		return b, nil
	}

	if k, ok := msg.(tea.KeyMsg); ok && key.Matches(k, keyActivate) {
		return b, b.cmd
	}

	return b, nil
}

// View renders the button using the focused, blurred, or disabled style.
func (b *Model) View() tea.View {
	return tea.NewView(b.StateStyles(b.Disabled(), b.Focused()).Render(b.label))
}

// Hint returns the button's hint text.
func (b *Model) Hint() string {
	return b.HintText
}
