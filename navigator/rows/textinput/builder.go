package textinput

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// Builder configures a text input row through chained method calls.
type Builder struct {
	model *Model
}

// NewBuilder creates a text input row builder with defaults applied.
func NewBuilder() *Builder {
	input := textinput.New()
	input.Prompt = ""

	m := &Model{input: input}
	m.SetStyles(DefaultStyles())

	return &Builder{
		model: m,
	}
}

// Label sets the row label.
func (b *Builder) Label(label string) *Builder {
	b.model.SetLabel(label)

	return b
}

// Placeholder sets the placeholder text shown when the input is empty.
func (b *Builder) Placeholder(placeholder string) *Builder {
	b.model.input.Placeholder = placeholder

	return b
}

// Width sets the width of the underlying input.
func (b *Builder) Width(width int) *Builder {
	b.model.SetWidth(width)

	return b
}

// Filter restricts which key messages reach the underlying input.
func (b *Builder) Filter(filter func(tea.KeyMsg) bool) *Builder {
	b.model.filter = filter

	return b
}

// Validator sets a validator run against the current string value.
func (b *Builder) Validator(fn func(string) error) *Builder {
	b.model.validator = fn

	return b
}

// EchoMode sets the echo mode of the underlying input.
func (b *Builder) EchoMode(mode textinput.EchoMode) *Builder {
	b.model.input.EchoMode = mode

	return b
}

// Styles replaces the default styles.
func (b *Builder) Styles(styles Styles) *Builder {
	b.model.SetStyles(styles)

	return b
}

// Build returns the configured text input row.
func (b *Builder) Build() *Model {
	return b.model
}
