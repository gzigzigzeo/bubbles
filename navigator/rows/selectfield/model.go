// Package selectfield provides a navigator row that lets the user pick one
// option from a fixed list. Enter or Space opens an inline dropdown; ↑/↓
// navigate; Enter commits; Esc cancels.
package selectfield

import (
	"errors"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator"
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
	picker     *navigator.Model
	validator  func(T) error
}

// pickerRow is a single focusable row inside the inline dropdown. It carries no
// activation message; selectfield handles the commit key itself.
type pickerRow[T comparable] struct {
	row.FocusedState

	label  string
	styles Styles
}

// Init satisfies [tea.Model].
func (r *pickerRow[T]) Init() tea.Cmd {
	return nil
}

// Update satisfies [tea.Model]. The picker navigator handles navigation keys,
// so the row does not need to interpret keys itself.
func (r *pickerRow[T]) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return r, nil
}

// View renders the cursor indicator and the option label.
func (r *pickerRow[T]) View() tea.View {
	variant := r.styles.Blurred.Picker
	cursor := lipgloss.NewStyle().Width(lipgloss.Width(variant.Cursor.Render())).Render("")

	if r.Focused() {
		variant = r.styles.Focused.Picker
		cursor = variant.Cursor.Render()
	}

	line := lipgloss.JoinHorizontal(lipgloss.Top, cursor, variant.Item.Render(r.label))

	return tea.NewView(line)
}

// New creates a select field with the given options, applying opts in order.
func New[T comparable](options []Option[T], opts ...FieldOption[T]) *Model[T] {
	m := &Model[T]{
		options: options,
	}

	pickerRows := make([]tea.Model, len(options))
	for i, o := range options {
		pickerRows[i] = &pickerRow[T]{label: o.Label}
	}

	m.picker = navigator.New(pickerRows...)

	if len(options) > 0 {
		m.committed = options[0].Value
	}

	m.SetStyles(DefaultStyles())

	for _, opt := range opts {
		opt(m)
	}

	return m
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

// SetStyles replaces the current styles and pushes them into the picker rows.
func (m *Model[T]) SetStyles(styles Styles) {
	m.StatefulStyles.SetStyles(styles)

	for _, r := range m.picker.Items() {
		if pr, ok := r.(*pickerRow[T]); ok {
			pr.styles = styles
		}
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

// Init initializes the picker.
func (m *Model[T]) Init() tea.Cmd {
	return m.picker.Init()
}

// Update handles opening the picker, navigating it, committing a value, and
// cancelling.
func (m *Model[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Disabled() {
			return m, nil
		}

		if m.pickerOpen {
			return m.updatePicker(msg)
		}

		if m.Focused() && key.Matches(msg, keyOpen) {
			m.openPicker()
		}
	}

	return m, nil
}

// openPicker enters picker mode, placing the cursor on the committed item.
func (m *Model[T]) openPicker() {
	m.pickerOpen = true
	m.picker.FocusFirst()

	if idx := m.indexOf(m.committed); idx >= 0 {
		m.picker.FocusIndex(idx)
	}
}

// updatePicker handles key input while the picker is open.
func (m *Model[T]) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keyEscape):
		m.pickerOpen = false

		return m, nil
	case key.Matches(msg, keySelect):
		if idx := m.picker.FocusedIndex(); idx >= 0 {
			m.committed = m.options[idx].Value
			_ = m.Validate()
		}
		m.pickerOpen = false

		return m, nil
	}

	updated, cmd := m.picker.Update(msg)
	m.picker = updated.(*navigator.Model)

	return m, cmd
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

// View renders the inline collapsed value when the picker is closed, or the
// inline value followed by the dropdown rows when open.
func (m *Model[T]) View() tea.View {
	if !m.pickerOpen {
		return tea.NewView(m.inlineView())
	}

	picker := m.picker.View().Content

	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, m.inlineView(), picker))
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

	if m.Focused() {
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

// CursorLine returns the line within View() holding the highlighted picker row,
// or 0 when the picker is closed. Implements [row.CursorAware].
func (m *Model[T]) CursorLine() int {
	if !m.pickerOpen {
		return 0
	}

	return 1 + m.picker.CursorLine()
}

// IsAtFirstFocusable reports whether the picker is closed, so the outer
// navigator can move focus up. Implements [row.BoundaryAware].
func (m *Model[T]) IsAtFirstFocusable() bool {
	return !m.pickerOpen
}

// IsAtLastFocusable reports whether the picker is closed, so the outer
// navigator can move focus down. Implements [row.BoundaryAware].
func (m *Model[T]) IsAtLastFocusable() bool {
	return !m.pickerOpen
}

// FocusFirst focuses the closed row. Implements [row.FocusReceiver].
func (m *Model[T]) FocusFirst() tea.Cmd {
	return m.Focus()
}

// FocusLast focuses the closed row. Implements [row.FocusReceiver].
func (m *Model[T]) FocusLast() tea.Cmd {
	return m.Focus()
}
