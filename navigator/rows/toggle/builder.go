package toggle

import (
	"charm.land/bubbles/v2/key"

	"github.com/gzigzigzeo/bubbles/navigator/row"
)

// Builder configures a toggle row through chained method calls.
type Builder struct {
	model *Model
}

// NewBuilder creates a toggle row builder with defaults applied.
func NewBuilder() *Builder {
	m := &Model{
		toggleKey: defaultToggleKey,
	}
	m.SetLabel("")
	m.SetStyles(DefaultStyles("On", "Off"))

	return &Builder{
		model: m,
	}
}

// Label sets the row label.
func (b *Builder) Label(label string) *Builder {
	b.model.SetLabel(label)

	return b
}

// Value sets the initial value.
func (b *Builder) Value(value bool) *Builder {
	b.model.value = value

	return b
}

// ToggleKeys sets the key binding that toggles the row.
func (b *Builder) ToggleKeys(keys ...string) *Builder {
	b.model.toggleKey = key.NewBinding(key.WithKeys(keys...), key.WithHelp(keys[0], "toggle"))

	return b
}

// Styles replaces the default styles.
func (b *Builder) Styles(styles Styles) *Builder {
	b.model.SetStyles(styles)

	return b
}

// DisableController creates a controller that disables targets when this
// toggle becomes true and enables them when it becomes false.
func (b *Builder) DisableController(targets ...row.Disableable) *DisableController {
	return NewDisableControllerFor(b.model, targets...)
}

// Build returns the configured toggle row.
func (b *Builder) Build() *Model {
	return b.model
}
