// Package selectfield provides a navigator row that lets the user pick one
// option from a fixed list. Enter or Space opens an inline dropdown; ↑/↓
// navigate; Enter commits; Esc cancels.
package selectfield

import (
	"errors"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator/row"
)

// ErrInvalidValue is a generic validation error that WithValidator callbacks
// can return.
var ErrInvalidValue = errors.New("invalid value")

var (
	keyOpen = key.NewBinding(
		key.WithKeys("enter", "space"),
		key.WithHelp("enter/space", "open"),
	)

	keySelect = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	)

	keyEscape = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	)
)

// Option is a label/value pair stored in a Model.
type Option[T comparable] struct {
	Value       T
	Label       string
	Description string
}

// FieldOption configures a Model at construction time. Named FieldOption
// (not Option) because Option[T] already names a selectable value/label pair.
type FieldOption[T comparable] func(*Model[T])

// WithLabel sets the field label.
func WithLabel[T comparable](label string) FieldOption[T] {
	return func(m *Model[T]) {
		m.SetLabel(label)
	}
}

// WithValidator sets a validator run against the committed value.
func WithValidator[T comparable](fn func(T) error) FieldOption[T] {
	return func(m *Model[T]) {
		m.validator = fn
	}
}

// Model is a navigator row for picking one option from a fixed list.
type Model[T comparable] struct {
	row.StatefulStyles[SelectStyles]
	row.FocusedState
	row.DisabledState
	row.LabelState
	row.ErrorState

	options    []Option[T]
	committed  T
	pickerOpen bool
	controller *Controller[T]
	validator  func(T) error
}

// New creates a select field with the given options, applying opts in order.
func New[T comparable](options []Option[T], opts ...FieldOption[T]) *Model[T] {
	b := NewBuilder[T]().Options(options)

	for _, opt := range opts {
		opt(b.model)
	}

	return b.Build()
}

// NewFromStrings creates a Model[string] where each string is both value and
// label, applying opts in order.
func NewFromStrings(options []string, opts ...FieldOption[string]) *Model[string] {
	converted := make([]Option[string], len(options))
	for i, s := range options {
		converted[i] = Option[string]{
			Value: s,
			Label: s,
		}
	}

	return New(converted, opts...)
}

// initController creates the controller that owns the picker rows.
func (m *Model[T]) initController() {
	m.controller = newController(m, m.options)
}

// Controller returns the select field's picker controller. The controller
// should be registered with the outer navigator via
// [navigator.Builder.WithControllerItems].
func (m *Model[T]) Controller() *Controller[T] {
	return m.controller
}

// SetStyles replaces the current styles and pushes them into the picker rows.
func (m *Model[T]) SetStyles(styles Styles) {
	m.StatefulStyles.SetStyles(styles)

	if m.controller != nil {
		m.controller.SetStyles(styles)
	}
}

// Get returns the currently selected option value.
func (m *Model[T]) Get() T {
	return m.committed
}

// Set selects the option whose Value equals value. No-op when value is not
// found.
func (m *Model[T]) Set(value T) {
	if idx := m.indexOf(value); idx >= 0 {
		m.committed = value
	}
}

// Validate runs the configured validator against the committed value and stores
// the result. It returns nil when no validator is set or the value is valid.
func (m *Model[T]) Validate() error {
	if m.validator == nil {
		m.SetErr(nil)

		return nil
	}

	err := m.validator(m.committed)
	m.SetErr(err)

	return err
}

// Init satisfies [tea.Model].
func (m *Model[T]) Init() tea.Cmd {
	return nil
}

// Update handles opening and cancelling the picker. Navigation, commit, and
// cancel while the picker is open are handled by the picker rows and the
// controller.
func (m *Model[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.Disabled() {
		return m, nil
	}

	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if m.pickerOpen {
		if key.Matches(km, keyEscape) {
			return m, m.controller.Close(false, *new(T))
		}

		// While the picker is opening the header row may still hold focus for a
		// single update. Consume navigation and selection keys so they do not
		// leave the picker region before the lock takes effect.
		return m, nil
	}

	if m.Focused() && key.Matches(km, keyOpen) {
		m.pickerOpen = true

		return m, m.controller.Open(m.committed)
	}

	return m, nil
}

// indexOf returns the index of the option whose Value equals value, or -1 if
// none.
func (m *Model[T]) indexOf(value T) int {
	for i, o := range m.options {
		if o.Value == value {
			return i
		}
	}

	return -1
}

// View renders the inline label and current selection. The picker option rows
// are rendered by the outer navigator as separate items.
func (m *Model[T]) View() tea.View {
	return tea.NewView(m.inlineView())
}

// inlineView renders the label and current selection with arrow brackets when
// focused.
func (m *Model[T]) inlineView() string {
	styles := m.StateStyles(m.Disabled(), m.Focused())

	labelStyle := styles.Label
	if m.Err() != nil {
		labelStyle = styles.Error
	}

	label := labelStyle.Render(m.Label())
	value := styles.Value.Render(m.committedLabel())

	if m.Disabled() {
		return lipgloss.JoinHorizontal(lipgloss.Top, label, value)
	}

	if m.Focused() || m.pickerOpen {
		arrows := lipgloss.JoinHorizontal(lipgloss.Top, styles.ArrowLeft.Render(), value, styles.ArrowRight.Render())

		return lipgloss.JoinHorizontal(lipgloss.Top, label, arrows)
	}

	leftPad := lipgloss.NewStyle().Width(lipgloss.Width(styles.ArrowLeft.Render())).Render("")
	rightPad := lipgloss.NewStyle().Width(lipgloss.Width(styles.ArrowRight.Render())).Render("")
	arrows := lipgloss.JoinHorizontal(lipgloss.Top, leftPad, value, rightPad)

	return lipgloss.JoinHorizontal(lipgloss.Top, label, arrows)
}

// committedLabel returns the Label of the committed option, or "" if there is
// none.
func (m *Model[T]) committedLabel() string {
	if idx := m.indexOf(m.committed); idx >= 0 {
		return m.options[idx].Label
	}

	return ""
}
