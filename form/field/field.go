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
	Blur()
	Focused() bool
}

// Disableable is implemented by entries that can be enabled and disabled.
type Disableable interface {
	Enable()
	Disable()
	Enabled() bool
	Disabled() bool
}

// KeyHinted is implemented by entries that claim keyboard bindings while focused.
type KeyHinted interface {
	Keys() []key.Binding
}

// Hinted is implemented by entries that show hint text while focused.
type Hinted interface {
	Hint() string
}

// FocusModel is implemented by composite Controls that manage focus over
// their own inner items via an embedded FocusState (e.g. buttonstack.Model).
// It lets an enclosing FocusState give the child first refusal on
// navigation keys (by comparing Position() before/after forwarding a
// message) and tells the child which end to preselect when focus enters
// from a given direction.
type FocusModel interface {
	Position() int
	SelectFirst()
	SelectLast()
}

// HeightAware is implemented by entries that can grow beyond one row
// (e.g. an open dropdown) and need a ceiling on how many rows to render.
type HeightAware interface {
	SetAvailableHeight(h int)
}

// LeftGutterAware is implemented by entries that render into the gutter
// column between the label and the entry (e.g. a scroll-position indicator).
type LeftGutterAware interface {
	// LeftGutter returns gutter content: one line per line of View().Content.
	// Return "" for lines with nothing to show.
	LeftGutter() string
}

// CursorAware is implemented by entries with internally navigable content
// (e.g. an open dropdown) so a form can keep the active line in view.
type CursorAware interface {
	// CursorLine returns the 0-indexed line within View() that should stay
	// visible - e.g. the currently highlighted dropdown option.
	CursorLine() int
}

// Validateable is implemented by entries that validate their own value and
// record, clear, or report the resulting error.
type Validateable interface {
	Validate() string
	Err() string
	SetErr(msg string)
}

// Sizeable is implemented by entries that participate in the form's shared
// label-column layout.
type Sizeable interface {
	Label() string
	ValueLeftPadding() int
	SetWidth(width int)
	SetLayout(labelWidth, maxValuePadding int)
	SetRowStyles(s Styles)

	// Unwrap returns the entry's wrapped Field[T] for optional capability
	// type-assertions Sizeable doesn't declare.
	Unwrap() AnyField
}

// Control is the structural interface a focusable, disableable widget must
// implement to participate in a form or a buttonstack.Model.
type Control interface {
	tea.Model
	Focusable
	Disableable
	KeyHinted
}

// Field is the structural interface a value-bearing widget implements, e.g.
// a text input or a select: Control plus width/padding metadata and a typed value.
type Field[T any] interface {
	Control

	// ValueLeftPadding returns the field value's left indent column count.
	ValueLeftPadding() int

	SetWidth(width int)

	Get() T
	Set(value T)
}

// AnyField is the type-erased shape of Field[T], used by Sizeable to
// promote layout-relevant methods without knowing T.
type AnyField interface {
	Control

	// ValueLeftPadding returns the field value's left indent column count.
	ValueLeftPadding() int

	SetWidth(width int)
}
