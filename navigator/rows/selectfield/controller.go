package selectfield

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/navigator"
	"github.com/gzigzigzeo/bubbles/navigator/row"
)

// Controller owns the picker option rows. It is registered with the outer
// navigator through [Model.Controller] so the picker rows are part of the flat
// focus list.
type Controller[T comparable] struct {
	header  *Model[T]
	options []Option[T]
	rows    []*pickerRow[T]
}

// newController creates a controller for the given header model and options.
// Picker rows start disabled so the outer navigator skips them when the picker
// is closed.
func newController[T comparable](header *Model[T], options []Option[T]) *Controller[T] {
	c := &Controller[T]{
		header:  header,
		options: options,
		rows:    make([]*pickerRow[T], len(options)),
	}

	for i, o := range options {
		c.rows[i] = &pickerRow[T]{
			header: header,
			value:  o.Value,
			label:  o.Label,
		}
		_ = c.rows[i].Disable()
	}

	c.SetStyles(header.Styles())

	return c
}

// Items returns the picker rows as tea models for the outer navigator.
func (c *Controller[T]) Items() []tea.Model {
	items := make([]tea.Model, len(c.rows))
	for i, r := range c.rows {
		items[i] = r
	}

	return items
}

// Open enables the picker rows and returns a command that asks the outer
// navigator to lock focus inside the picker and focus the committed option.
func (c *Controller[T]) Open(committed T) tea.Cmd {
	for _, r := range c.rows {
		_ = r.Enable()
	}

	var focus tea.Model
	for i, o := range c.options {
		if o.Value == committed {
			focus = c.rows[i]

			break
		}
	}

	return func() tea.Msg {
		return navigator.LockFocusMsg{
			Range: c.Items(),
			Focus: focus,
		}
	}
}

// Close disables the picker rows, updates the header state, and returns a
// command that asks the outer navigator to unlock focus and return focus to
// the header. When commit is true, value becomes the header's committed value.
func (c *Controller[T]) Close(commit bool, value T) tea.Cmd {
	c.header.pickerOpen = false

	if commit {
		c.header.committed = value
		_ = c.header.Validate()
	}

	for _, r := range c.rows {
		_ = r.Disable()
	}

	return func() tea.Msg {
		return navigator.UnlockFocusMsg{
			Focus: c.header,
		}
	}
}

// Update is part of the navigator controller interface. The picker rows close
// the picker directly, so the controller has no messages to handle.
func (c *Controller[T]) Update(_ tea.Msg) tea.Cmd {
	return nil
}

// SetStyles pushes styles into every picker row.
func (c *Controller[T]) SetStyles(styles Styles) {
	for _, r := range c.rows {
		r.SetStyles(styles)
	}
}

// pickerRow is a single focusable option inside the inline dropdown.
type pickerRow[T comparable] struct {
	row.StatefulStyles[SelectStyles]
	row.FocusedState
	row.DisabledState

	header *Model[T]
	value  T
	label  string
}

// Init satisfies [tea.Model].
func (r *pickerRow[T]) Init() tea.Cmd {
	return nil
}

// Update commits or cancels picker navigation when the row is focused and
// enabled. The picker row talks directly to its header controller so the close
// happens synchronously without an extra message round-trip.
func (r *pickerRow[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !r.Focused() || r.Disabled() {
		return r, nil
	}

	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return r, nil
	}

	if key.Matches(km, keySelect) {
		return r, r.header.controller.Close(true, r.value)
	}

	if key.Matches(km, keyEscape) {
		var zero T

		return r, r.header.controller.Close(false, zero)
	}

	return r, nil
}

// View renders the cursor indicator and option label when enabled. Disabled
// rows render an empty string so they do not appear while the picker is closed.
func (r *pickerRow[T]) View() tea.View {
	if r.Disabled() {
		return tea.NewView("")
	}

	styles := r.StateStyles(false, r.Focused())
	cursor := styles.Picker.Cursor.Render()

	if !r.Focused() {
		cursor = lipgloss.NewStyle().Width(lipgloss.Width(cursor)).Render("")
	}

	line := lipgloss.JoinHorizontal(lipgloss.Top, cursor, styles.Picker.Item.Render(r.label))

	return tea.NewView(line)
}

// Focus marks the row as focused.
func (r *pickerRow[T]) Focus() tea.Cmd {
	r.FocusedState.Focus()

	return nil
}

// Blur removes focus from the row.
func (r *pickerRow[T]) Blur() tea.Cmd {
	r.FocusedState.Blur()

	return nil
}

// Focused reports whether the row is focused.
func (r *pickerRow[T]) Focused() bool {
	return r.FocusedState.Focused()
}
