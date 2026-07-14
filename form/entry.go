// Package form2 defines form rows as plain tea.Model values that optionally
// implement field.Control, Sizeable, field.Validateable, and/or field.Hinted.
package form

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/form/field"
)

// Sizeable is implemented by entries that participate in the form's shared
// label-column layout.
type Sizeable interface {
	Label() string
	SetWidth(width int)
	SetLayout(labelWidth int)

	// Unwrap returns the entry's wrapped Field[T] for optional capability
	// type-assertions Sizeable doesn't declare.
	Unwrap() field.AnyField
}

// hintState holds hint text, implementing field.Hinted and the private
// hintSetter constraint. Embedded by FieldEntry.
type hintState struct {
	hint string
}

// Hint returns the entry's hint text.
func (h *hintState) Hint() string {
	return h.hint
}

// setHint sets the entry's hint text. Used by the shared WithHint option.
func (h *hintState) setHint(hint string) {
	h.hint = hint
}

// FieldEntry wraps a field.Field[T] with label, hint, and validator
// metadata, and renders its own decorated row.
type FieldEntry[T any] struct {
	field.Field[T]
	hintState

	label      string
	err        error
	validator  func(T) string
	labelWidth int
	styles     Styles
}

// NewField creates a FieldEntry labeled label for f, applying opts in order.
func NewField[T any](label string, f field.Field[T], opts ...FieldOption[T]) *FieldEntry[T] {
	e := &FieldEntry[T]{
		Field: f,
		label: label,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Validate calls the entry's validator with the field's current value,
// returning "" when there's no validator or the field is disabled.
func (e *FieldEntry[T]) Validate() string {
	if e.validator == nil || e.Disabled() {
		return ""
	}

	return e.validator(e.Get())
}

// Label returns the entry's label.
func (e *FieldEntry[T]) Label() string {
	return e.label
}

// Err returns the entry's current validation error, set by Model.Validate.
func (e *FieldEntry[T]) Err() error {
	return e.err
}

// SetErr sets the entry's current validation error. Used by Model.Validate.
func (e *FieldEntry[T]) SetErr(err error) {
	e.err = err
}

// Unwrap returns the entry's wrapped field.Field[T] as a field.AnyField.
func (e *FieldEntry[T]) Unwrap() field.AnyField {
	return e.Field
}

// SetLayout stores the label width Model computed across all sizeable entries.
func (e *FieldEntry[T]) SetLayout(labelWidth int) {
	e.labelWidth = labelWidth
}

// SetRowStyles stores the row-chrome styles (cursor, label, gutter, error)
// Model pushes down whenever its own styles change.
func (e *FieldEntry[T]) SetRowStyles(s Styles) {
	e.styles = s
}

// View renders this entry's full row: cursor, aligned label, gutter,
// wrapped field, and optional error.
func (e *FieldEntry[T]) View() tea.View {
	var (
		cursor     lipgloss.Style
		labelStyle lipgloss.Style
	)

	switch {
	case e.Disabled():
		cursor = e.styles.CursorBlurred
		labelStyle = e.styles.LabelDisabled
	case e.Focused():
		cursor = e.styles.CursorFocused
		labelStyle = e.styles.LabelFocused
	default:
		cursor = e.styles.CursorBlurred
		labelStyle = e.styles.LabelBlurred
	}

	cursorStr := cursor.String()
	cursorW := lipgloss.Width(cursorStr)

	label := labelStyle.Width(e.labelWidth - cursorW).Render(e.label)

	gutter := e.renderGutter()
	parts := []string{cursorStr, label, gutter, e.Field.View().Content}

	if e.err != nil {
		parts = append(parts, e.styles.ErrText.Render(" "+e.err.Error()))
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, parts...)

	return tea.NewView(e.styles.Row.Render(content))
}

// renderGutter renders the empty gutter column when the field does not own
// the gutter, or returns "" when the field renders the gutter itself.
func (e *FieldEntry[T]) renderGutter() string {
	if go_, ok := e.Unwrap().(field.GutterOwner); ok && go_.OwnsGutter() {
		return ""
	}

	return e.styles.EmptyGutter.Render("")
}
