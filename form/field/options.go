package field

// styleSetter is implemented by widget types whose styles can be replaced wholesale.
type styleSetter[S any] interface {
	SetStyles(styles S)
}

// widthSetter is implemented by widget types with an adjustable display width.
type widthSetter interface {
	SetWidth(width int)
}

// WithStyles returns an option that calls SetStyles(s) on construction, for
// any widget type implementing Styler.
func WithStyles[F styleSetter[S], S any](s S) func(F) {
	return func(f F) {
		f.SetStyles(s)
	}
}

// WithWidth returns an option that calls SetWidth(w) on construction, for
// any widget type implementing Widther.
func WithWidth[F widthSetter](w int) func(F) {
	return func(f F) {
		f.SetWidth(w)
	}
}
