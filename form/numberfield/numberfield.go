// Package numberfield provides a form2 field for an integer value, rendered
// as a text input that only accepts numeric characters.
package numberfield

import (
	"strconv"

	tea "charm.land/bubbletea/v2"

	"github.com/gzigzigzeo/bubbles/form/textinputfield"
)

// Styles defines the styles for the Model component.
type Styles = textinputfield.Styles

// Model is a form field that represents an integer value, rendered as a text input
// that only accepts numeric characters.
type Model struct {
	textinputfield.State
}

// Option configures a Model at construction time.
type Option func(*Model)

// New creates a new Model, applying opts in order.
func New(opts ...Option) *Model {
	nf := &Model{}
	nf.State = textinputfield.New(textinputfield.WithKeyFilter(isNumericKey))

	for _, opt := range opts {
		opt(nf)
	}

	return nf
}

// Get returns the current value of the Model as an integer.
func (f *Model) Get() int {
	v, _ := strconv.Atoi(f.Input().Value())

	return v
}

// Set sets the value of the Model to the provided integer.
func (f *Model) Set(v int) {
	f.Input().SetValue(strconv.Itoa(v))
}

// Update forwards msg to the underlying text input, filtering out non-numeric input.
func (f *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return f, f.State.Update(msg)
}

// isNumericKey checks if the provided key message corresponds to a numeric character.
func isNumericKey(k tea.KeyMsg) bool {
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
