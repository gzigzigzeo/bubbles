// Package field defines the capability interfaces and composable state
// mixins shared by every form2 widget.
package field

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// Focusable is implemented by entries that can hold and release keyboard focus.
type Focusable interface {
	Focus() tea.Cmd
	Blur() tea.Cmd
	Focused() bool
}

// Disableable is implemented by entries that can be enabled and disabled.
type Disableable interface {
	Enable() tea.Cmd
	Disable() tea.Cmd
	Disabled() bool
}

// Hinted is implemented by entries that show hint text and claim keyboard
// bindings while focused.
type Hinted interface {
	Hint() string
	Keys() []key.Binding
}

// FocusModel is implemented by composite Controls that manage focus over
// their own inner items via an embedded FocusState (e.g. buttonstack.Model).
type FocusModel interface {
	Position() int // TODO: I think, it should be removed
	FocusFirst() tea.Cmd
	FocusLast() tea.Cmd
}

// HeightAware is implemented by entries that can grow beyond one row
// (e.g. an open dropdown) and need a ceiling on how many rows to render.
type HeightAware interface {
	SetAvailableHeight(h int)
}

// GutterOwner is implemented by fields that render the gutter column as part
// of their own View content. When OwnsGutter returns true, the form skips its
// own gutter column and allocates gutter width to the field via SetWidth.
type GutterOwner interface {
	OwnsGutter() bool
}

// CursorAware is implemented by entries with internally navigable content
// (e.g. an open dropdown) so a form can keep the active line in view.
type CursorAware interface {
	CursorLine() int
}

// Validateable is implemented by entries that validate their own value and
// record, clear, or report the resulting error.
type Validateable interface {
	Validate() string
	Err() error
	SetErr(err error)
}

// Control is the structural interface a focusable, disableable widget must
// implement to participate in a form or a buttonstack.Model.
type Control interface {
	tea.Model
	Focusable
	Disableable
	Keys() []key.Binding
}

// Field is the structural interface a value-bearing widget implements, e.g.
// a text input or a select: Control plus width/padding metadata and a typed value.
type Field[T any] interface {
	Control

	SetWidth(width int)

	Get() T
	Set(value T)
}

// AnyField is the type-erased shape of Field[T], used by Sizeable to
// promote layout-relevant methods without knowing T.
type AnyField interface {
	Control

	SetWidth(width int)
}
