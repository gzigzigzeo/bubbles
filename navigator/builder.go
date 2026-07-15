package navigator

import tea "charm.land/bubbletea/v2"

// Builder configures a navigator through chained method calls.
type Builder struct {
	items       []tea.Model
	controllers []Controller
	wrap        bool
	height      int
	viewport    Viewport
}

// NewBuilder creates a navigator builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// WithItems appends raw rows to the navigator.
func (b *Builder) WithItems(items ...tea.Model) *Builder {
	b.items = append(b.items, items...)

	return b
}

// WithControllerItems adds the rows owned by a controller to the navigator and
// registers the controller for global message handling.
func (b *Builder) WithControllerItems(c itemController) *Builder {
	b.items = append(b.items, c.Items()...)
	b.controllers = append(b.controllers, c)

	return b
}

// WithController registers a controller that has no rows of its own. The
// navigator will call Update on it for every message.
func (b *Builder) WithController(c Controller) *Builder {
	b.controllers = append(b.controllers, c)

	return b
}

// WithWrap enables wrap-at-boundaries focus movement.
func (b *Builder) WithWrap() *Builder {
	b.wrap = true

	return b
}

// WithHeight sets the viewport height.
func (b *Builder) WithHeight(height int) *Builder {
	b.height = height

	return b
}

// WithViewport attaches a viewport.
func (b *Builder) WithViewport(v Viewport) *Builder {
	b.viewport = v

	return b
}

// Build returns the configured navigator model.
func (b *Builder) Build() *Model {
	nav := New(b.items...)

	if b.wrap {
		nav.Wrap()
	}

	if b.viewport != nil {
		nav.ViewportController().SetViewport(b.viewport)
	}

	if b.height > 0 {
		nav.ViewportController().SetHeight(b.height)
	}

	nav.controllers = b.controllers

	return nav
}
