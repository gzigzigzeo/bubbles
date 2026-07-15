package selectfield

// Builder configures a select field through chained method calls.
type Builder[T comparable] struct {
	model *Model[T]
}

// NewBuilder creates a select field builder with defaults applied.
func NewBuilder[T comparable]() *Builder[T] {
	m := &Model[T]{}
	m.SetLabel("")
	m.SetStyles(DefaultStyles())

	return &Builder[T]{
		model: m,
	}
}

// Options sets the selectable options.
func (b *Builder[T]) Options(options []Option[T]) *Builder[T] {
	b.model.options = options

	if len(options) > 0 {
		b.model.committed = options[0].Value
	}

	return b
}

// Label sets the field label.
func (b *Builder[T]) Label(label string) *Builder[T] {
	b.model.SetLabel(label)

	return b
}

// Validator sets a validator run against the committed value.
func (b *Builder[T]) Validator(fn func(T) error) *Builder[T] {
	b.model.validator = fn

	return b
}

// Styles replaces the default styles.
func (b *Builder[T]) Styles(styles Styles) *Builder[T] {
	b.model.SetStyles(styles)

	return b
}

// Build returns the configured select field with its controller initialized.
func (b *Builder[T]) Build() *Model[T] {
	b.model.initController()

	return b.model
}
