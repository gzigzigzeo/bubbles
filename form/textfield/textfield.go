// Package textfield provides a form2 field for a string value, rendered as a text input.
package textfield

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/gzigzigzeo/bubbles/form/textinputfield"
)

// Styles defines the styles for the Model component.
type Styles = textinputfield.Styles

// Model is a form field that represents a string value, rendered as a text input.
type Model struct {
	textinputfield.State
}

// Option configures a Model at construction time.
type Option func(*Model)

// WithPlaceholder sets the placeholder text shown when the field is empty.
func WithPlaceholder(s string) Option {
	return func(f *Model) {
		f.SetPlaceholder(s)
	}
}

// WithEchoMode sets the echo mode of the underlying text input (e.g. textinput.EchoPassword).
func WithEchoMode(mode textinput.EchoMode) Option {
	return func(f *Model) {
		f.SetEchoMode(mode)
	}
}

// New creates a new Model, applying opts in order.
func New(opts ...Option) *Model {
	tf := &Model{}
	tf.State = textinputfield.New()

	for _, opt := range opts {
		opt(tf)
	}

	return tf
}

// SetPlaceholder sets the placeholder text shown when the Model is empty.
func (f *Model) SetPlaceholder(s string) {
	f.Input().Placeholder = s
}

// SetEchoMode sets the echo mode of the underlying text input (e.g. textinput.EchoPassword).
func (f *Model) SetEchoMode(mode textinput.EchoMode) {
	f.Input().EchoMode = mode
}

// Get returns the current value of the Model as a string.
func (f *Model) Get() string {
	return f.Input().Value()
}

// Set sets the value of the Model to the provided string.
func (f *Model) Set(s string) {
	f.Input().SetValue(s)
}

// Update forwards msg to the underlying text input.
func (f *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return f, f.State.Update(msg)
}
