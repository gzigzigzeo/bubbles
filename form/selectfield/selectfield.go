package selectfield

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/form/field"
	"github.com/gzigzigzeo/bubbles/menu"
)

var keySelect = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "select"),
)

// pickerMarginTop/pickerMarginBottom are blank lines above and below the open dropdown.
const (
	pickerMarginTop    = 1
	pickerMarginBottom = 1
)

// Option is a label/value pair stored in a Model.
type Option[T comparable] struct {
	Value       T
	Label       string
	Description string // optional; passed through to the picker's description column
}

var keyOpen = key.NewBinding(
	key.WithKeys("enter", "space"),
	key.WithHelp("enter", "open"),
)

var keyEscape = key.NewBinding(
	key.WithKeys("esc"),
	key.WithHelp("esc", "cancel"),
)

// Model is a form field for picking one option from a fixed list.
// Enter/Space opens an inline dropdown; ↑/↓ navigate; Enter commits; Esc cancels.
type Model[T comparable] struct {
	field.StyledState[SelectStyles]
	field.FocusedState
	field.DisabledState
	field.NoopInit

	options []Option[T]
	menu    *menu.Model[T]

	committed       T
	width           int
	availableHeight int

	pickerOpen bool
}

// SetStyles replaces the current styles and pushes the active picker variant into the menu.
func (f *Model[T]) SetStyles(styles Styles) {
	f.StyledState.SetStyles(styles)
	f.menu.SetStyles(f.StateStyles(f.Disabled(), f.Focused()).Picker)
}

// Enable marks the field as enabled and pushes the active picker variant into the menu.
func (f *Model[T]) Enable() tea.Cmd {
	f.DisabledState.Enable()
	f.menu.SetStyles(f.StateStyles(f.Disabled(), f.Focused()).Picker)

	return nil
}

// Disable marks the field as disabled and pushes the active picker variant into the menu.
func (f *Model[T]) Disable() tea.Cmd {
	f.DisabledState.Disable()
	f.menu.SetStyles(f.StateStyles(f.Disabled(), f.Focused()).Picker)

	return nil
}

// Focus marks the field as focused and pushes the focused picker variant into the menu.
func (f *Model[T]) Focus() tea.Cmd {
	cmd := f.FocusedState.Focus()
	f.menu.SetStyles(f.StateStyles(f.Disabled(), true).Picker)

	return cmd
}

// Blur marks the field as blurred and pushes the blurred picker variant into the menu.
func (f *Model[T]) Blur() tea.Cmd {
	f.FocusedState.Blur()
	f.menu.SetStyles(f.StateStyles(f.Disabled(), false).Picker)

	return nil
}

// SetAvailableHeight sets a ceiling on how many dropdown rows this field may render.
func (f *Model[T]) SetAvailableHeight(h int) {
	f.availableHeight = h

	if f.pickerOpen {
		f.menu.SetHeight(f.pickerVisible())
	}
}

// pickerVisible returns how many dropdown rows to show given the available height.
func (f *Model[T]) pickerVisible() int {
	n := len(f.options)
	overhead := 1 + pickerMarginTop + pickerMarginBottom // inline row + blank lines around the dropdown
	budget := f.availableHeight - overhead

	if budget < n {
		return max(1, budget)
	}

	return n
}

// toMenuOptions converts a Model's Value/Label options into the
// Value/Name shape menu.Model expects.
func toMenuOptions[T comparable](options []Option[T]) []menu.Option[T] {
	opts := make([]menu.Option[T], len(options))
	for i, o := range options {
		opts[i] = menu.Option[T]{Name: o.Label, Description: o.Description, Value: o.Value}
	}

	return opts
}

// FieldOption configures a Model at construction time. Named
// FieldOption (not Option) because Option[T] already names a selectable
// value/label pair.
type FieldOption[T comparable] func(*Model[T])

// New creates a Model with the given options, applying opts in order.
func New[T comparable](options []Option[T], opts ...FieldOption[T]) *Model[T] {
	f := &Model[T]{
		options: options,
		menu:    menu.New(toMenuOptions(options)),
	}

	if len(options) > 0 {
		f.committed = options[0].Value
		f.menu.SetMarker(f.committed)
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

// NewFromStrings creates a Model[string] where each string is both
// value and label, applying opts in order.
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

// SetOptions replaces the field's options, e.g. when a dependent field
// changes what should be selectable. The committed value stays if still
// present, otherwise it resets to the first new option (or the zero value).
func (f *Model[T]) SetOptions(options []Option[T]) {
	f.options = options
	f.menu.SetOptions(toMenuOptions(options))

	for _, o := range options {
		if o.Value == f.committed {
			f.menu.SetMarker(f.committed)
			return
		}
	}

	var zero T
	f.committed = zero

	if len(options) > 0 {
		f.committed = options[0].Value
	}

	f.menu.SetMarker(f.committed)
}

// Get returns the currently selected option value.
func (f *Model[T]) Get() T {
	return f.committed
}

// Set selects the option whose Value equals value. No-op when value is not found.
func (f *Model[T]) Set(value T) {
	for _, o := range f.options {
		if o.Value == value {
			f.committed = value
			f.menu.SetMarker(value)

			return
		}
	}
}

// OwnsGutter returns true: this field renders the gutter column as part of
// its own View content (scroll indicator on the left of each picker row,
// blank gutter-width prefix on the closed inline row).
func (f *Model[T]) OwnsGutter() bool {
	return true
}

// SetWidth stores the total allocated width (field + gutter) and passes it
// to the picker so picker rows fill the full slot.
func (f *Model[T]) SetWidth(w int) {
	f.width = w
	f.menu.SetWidth(w)
}

// scrollWidth returns the visual width of the picker's scroll-position
// indicator, which the field renders as the gutter prefix on every row.
func (f *Model[T]) scrollWidth() int {
	return lipgloss.Width(f.StateStyles(f.Disabled(), f.Focused()).Picker.ScrollUp.Render())
}

// Keys returns key bindings appropriate for the current state:
// the open-picker binding when closed, or navigate/select/cancel bindings when open.
func (f *Model[T]) Keys() []key.Binding {
	if f.Disabled() {
		return nil
	}

	if f.pickerOpen {
		return append(f.menu.Keys(), keyEscape)
	}

	return []key.Binding{keyOpen}
}

// Update handles picker navigation and opening the picker.
func (f *Model[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !f.Focused() || f.Disabled() || len(f.options) == 0 {
			return f, nil
		}

		if f.pickerOpen {
			return f, f.updatePicker(msg)
		}

		if key.Matches(msg, keyOpen) {
			f.openPicker()
		}
	}

	return f, nil
}

// openPicker enters picker mode, placing the cursor on the committed item.
func (f *Model[T]) openPicker() {
	f.pickerOpen = true
	f.menu.SetHeight(f.pickerVisible())
	f.menu.SetValue(f.committed)
}

// updatePicker handles key input while the picker is open: Select commits
// and closes it, Escape closes without committing, everything else is
// forwarded to the menu for navigation/scrolling.
func (f *Model[T]) updatePicker(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, keySelect):
		f.committed = f.menu.Cursor()
		f.menu.SetMarker(f.committed)
		f.pickerOpen = false

		return nil

	case key.Matches(msg, keyEscape):
		f.pickerOpen = false

		return nil
	}

	_, cmd := f.menu.Update(msg)

	return cmd
}

// View renders the inline collapsed value when the picker is closed, or the
// inline value followed by the dropdown rows when open. The field always
// prepends a gutter-width prefix (scroll indicator or blank) so the form
// does not need to render a separate gutter column (see OwnsGutter).
func (f *Model[T]) View() tea.View {
	gutterPrefix := "" //strings.Repeat(" ", f.scrollWidth())

	if !f.pickerOpen {
		return tea.NewView(gutterPrefix + f.inlineView())
	}

	// Blank lines around the dropdown replace Menu's removed
	// Styles.Container.Margin concept.
	top := strings.Repeat("\n", pickerMarginTop)
	bottom := strings.Repeat("\n", pickerMarginBottom)

	return tea.NewView(gutterPrefix + f.inlineView() + "\n" + top + f.menu.View().Content + bottom)
}

// CursorLine returns the 0-indexed line within View() holding the
// highlighted option (field.CursorAware), or 0 when the picker is closed.
func (f *Model[T]) CursorLine() int {
	if !f.pickerOpen {
		return 0
	}

	return 1 + pickerMarginTop + f.menu.CursorLine()
}

// inlineView renders the current selection with ◀ value ▶ brackets when focused.
// Blurred rows reserve the same space as the arrows with plain spaces so all
// values align at the same column regardless of focus state.
func (f *Model[T]) inlineView() string {
	s := f.StateStyles(f.Disabled(), f.Focused())

	label := f.committedLabel()

	if f.Disabled() {
		return s.Value.Render(label)
	}

	arrowL := s.ArrowLeft.Render()
	arrowR := s.ArrowRight.Render()
	value := s.Value.Render(label)

	if f.Focused() {
		return arrowL + value + arrowR
	}

	return strings.Repeat(" ", lipgloss.Width(arrowL)) + value + strings.Repeat(" ", lipgloss.Width(arrowR))
}

// Menu returns the underlying picker menu.Model, for callers that need direct
// access beyond what Model's own methods expose.
func (f *Model[T]) Menu() *menu.Model[T] {
	return f.menu
}

// committedLabel returns the Label of the option whose Value equals the
// committed value, or "" if there is none.
func (f *Model[T]) committedLabel() string {
	for _, o := range f.options {
		if o.Value == f.committed {
			return o.Label
		}
	}

	return ""
}
