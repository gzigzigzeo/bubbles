package form

import (
	tea "charm.land/bubbletea/v2"
)

// WithEntry appends m as an entry, in call order, classified by whichever
// optional capability interfaces it implements.
func WithEntry(m tea.Model) Option {
	return func(f *Model) {
		f.appendEntry(m)
	}
}

// WithStyles sets the form's chrome styles.
func WithStyles(s Styles) Option {
	return func(f *Model) {
		f.SetStyles(s)
	}
}

// WithWidth sets the form's content width.
func WithWidth(w int) Option {
	return func(f *Model) {
		f.SetWidth(w)
	}
}

// FieldOption configures a FieldEntry at construction time.
type FieldOption[T any] func(*FieldEntry[T])

// WithHint returns an option that sets an entry's hint text, shown while it
// is focused.
func WithHint[T any](hint string) FieldOption[T] {
	return func(e *FieldEntry[T]) {
		e.setHint(hint)
	}
}

// WithValidator sets a FieldEntry's validator, run by Validate.
func WithValidator[T any](validator func(T) string) FieldOption[T] {
	return func(e *FieldEntry[T]) {
		e.validator = validator
	}
}
