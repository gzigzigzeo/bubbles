package button

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Builder configures a single button row through chained method calls.
type Builder struct {
	model *Model
}

// NewBuilder creates a button row builder with defaults applied.
func NewBuilder() *Builder {
	m := &Model{
		pressKey: defaultPressKey,
	}
	m.SetLabel("")
	m.SetStyles(DefaultStyles())

	return &Builder{
		model: m,
	}
}

// Label sets the button label.
func (b *Builder) Label(label string) *Builder {
	b.model.SetLabel(label)

	return b
}

// Msg sets the message emitted when the button is pressed.
func (b *Builder) Msg(msg tea.Msg) *Builder {
	b.model.SetMsg(msg)

	return b
}

// PressKeys sets the key binding that presses the button.
func (b *Builder) PressKeys(keys ...string) *Builder {
	b.model.pressKey = newPressKey(keys...)

	return b
}

// Styles replaces the default styles.
func (b *Builder) Styles(styles Styles) *Builder {
	b.model.SetStyles(styles)

	return b
}

// Build returns the configured button row.
func (b *Builder) Build() *Model {
	return b.model
}

// StackBuilder configures a horizontal button stack through chained method
// calls.
type StackBuilder struct {
	buttons []tea.Model
	styles  StackStyles
}

// NewStackBuilder creates a button stack builder.
func NewStackBuilder() *StackBuilder {
	return &StackBuilder{
		buttons: nil,
		styles:  StackStyles{Wrapper: lipgloss.NewStyle()},
	}
}

// Add appends buttons to the stack.
func (b *StackBuilder) Add(buttons ...*Model) *StackBuilder {
	for _, btn := range buttons {
		b.buttons = append(b.buttons, btn)
	}

	return b
}

// WrapperStyle sets the wrapper style applied around the joined buttons.
func (b *StackBuilder) WrapperStyle(style lipgloss.Style) *StackBuilder {
	b.styles.Wrapper = style

	return b
}

// Build returns the configured button stack.
func (b *StackBuilder) Build() *Stack {
	stack := NewStack(b.buttons...)
	stack.SetStyles(b.styles)

	return stack
}

// newPressKey builds a press key binding for the given keys.
func newPressKey(keys ...string) key.Binding {
	return key.NewBinding(key.WithKeys(keys...), key.WithHelp(keys[0], "press"))
}
