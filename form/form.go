package form

import (
	"errors"
	"slices"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gzigzigzeo/bubbles/form/field"
)

var keyDown = key.NewBinding(
	key.WithKeys("down"),
	key.WithHelp("↓", "next"),
)

var keyUp = key.NewBinding(
	key.WithKeys("up"),
	key.WithHelp("↑", "prev"),
)

var keyEnter = key.NewBinding(key.WithKeys("enter"))

// Model manages an ordered list of entries with keyboard navigation,
// rendered as a vertical stack clipped by an inner viewport.
type Model struct {
	entries       []tea.Model          // every row, in call order; each renders via its own View()
	focus         field.FocusState     // focusable subset, in call order, with the current focus index
	focusableLine []int                // parallel to focus.Items(): entries-index each occupies, for scroll-into-view
	sizeable      []Sizeable           // subset that participates in label-column layout
	validatable   []field.Validateable // subset that participates in Validate()

	styles     Styles
	labelWidth int
	height     int
	width      int // last content width passed to SetWidth; 0 = not yet called

	vp viewport.Model
}

// Option configures a Model at construction time.
type Option func(*Model)

// appendEntry appends m to entries and registers it in whichever
// capability lists it qualifies for: focus, sizeable, and/or validatable.
func (f *Model) appendEntry(m tea.Model) {
	f.entries = append(f.entries, m)

	if _, ok := m.(field.Control); ok {
		f.focusableLine = append(f.focusableLine, len(f.entries)-1)
	}

	if s, ok := m.(Sizeable); ok {
		f.sizeable = append(f.sizeable, s)
	}

	if v, ok := m.(field.Validateable); ok {
		f.validatable = append(f.validatable, v)
	}
}

// New creates a Model with no entries, applying opts in order. The first
// non-disabled focusable entry receives initial focus.
func New(opts ...Option) *Model {
	f := &Model{vp: viewport.New()}

	// Restrict viewport keys to PgUp/PgDn; other defaults (space, j/k, etc.)
	// would steal keystrokes meant for text fields and selects.
	f.vp.KeyMap = viewport.KeyMap{
		PageDown: key.NewBinding(key.WithKeys("pgdown")),
		PageUp:   key.NewBinding(key.WithKeys("pgup")),
	}

	for _, opt := range opts {
		opt(f)
	}

	f.focus = field.NewFocusState(f.entries...)

	f.SetStyles(f.styles) // re-finalize styles (+ width, if set) regardless of opts order
	f.focus.FocusFirst()

	return f
}

// rowStylesSetter is implemented by sizeable entries that accept row chrome styles.
type rowStylesSetter interface {
	SetRowStyles(s Styles)
}

// SetStyles replaces the chrome styles, pushing them to every sizeable
// entry and redoing layout if SetWidth was already called.
func (f *Model) SetStyles(s Styles) {
	f.styles = s

	for _, e := range f.sizeable {
		if ss, ok := e.(rowStylesSetter); ok {
			ss.SetRowStyles(s)
		}
	}

	if f.width > 0 {
		f.applyWidth()
	}
}

// SetWidth sets the total content width and distributes it across label and field columns.
func (f *Model) SetWidth(contentWidth int) {
	f.width = contentWidth
	f.applyWidth()
}

// applyWidth performs width-dependent layout using the most-recently-set
// width and styles.
func (f *Model) applyWidth() {
	f.vp.SetWidth(f.width)

	if len(f.sizeable) == 0 {
		return
	}

	labelLens := make([]int, len(f.sizeable))

	for i, e := range f.sizeable {
		labelLens[i] = len([]rune(e.Label()))
	}

	f.labelWidth = slices.Max(labelLens) + 5

	gutterWidth := lipgloss.Width(f.styles.EmptyGutter.Render(""))
	fieldWidth := f.width - f.labelWidth - gutterWidth

	for _, e := range f.sizeable {
		w := fieldWidth
		if go_, ok := e.Unwrap().(field.GutterOwner); ok && go_.OwnsGutter() {
			w += gutterWidth
		}
		e.SetWidth(w)
		e.SetLayout(f.labelWidth)
	}
}

// SetHeight sets the total content height and sizes the inner viewport.
func (f *Model) SetHeight(h int) {
	f.height = h
	f.vp.SetHeight(max(1, h))
	f.syncFocusedHeight()
	f.refreshContent()
	f.syncScroll()
}

// syncFocusedHeight passes the form height as an available-height ceiling to
// the focused entry's field if it implements field.HeightAware.
func (f *Model) syncFocusedHeight() {
	c := f.focus.Current()
	if c == nil {
		return
	}

	s, ok := c.(Sizeable)
	if !ok {
		return
	}

	if ha, ok := s.Unwrap().(field.HeightAware); ok {
		ha.SetAvailableHeight(max(1, f.height))
	}
}

// Init initializes every entry and then focuses the initially-focused row.
func (f *Model) Init() tea.Cmd {
	return tea.Batch(field.Init(f.entries...), f.focus.Focus())
}

// Update handles navigation and forwards messages to the focused row,
// then scrolls the viewport to keep the focused row visible.
func (f *Model) Update(msg tea.Msg) tea.Cmd {
	f.syncFocusedHeight()

	var vpCmd tea.Cmd

	f.vp, vpCmd = f.vp.Update(msg)

	cmd := f.updateRows(msg)

	// Refresh content before scrolling: syncScroll's SetYOffset is clamped
	// against the viewport's currently-stored content, which is otherwise
	// only refreshed by View() - one render late. Without this, a scroll
	// target computed off this row's freshly-changed height gets silently
	// clamped against the old (shorter) content and only takes effect on
	// the next keypress, reading as a one-line-delayed scroll jump.
	f.refreshContent()
	f.syncScroll()

	return tea.Batch(vpCmd, cmd)
}

// updateRows routes navigation keys and forwards other messages to the focused
// row; WindowSizeMsg is broadcast to all entries.
func (f *Model) updateRows(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmds := make([]tea.Cmd, 0, len(f.entries))

		for _, m := range f.entries {
			if _, cmd := m.Update(msg); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

		return tea.Batch(cmds...)

	case tea.KeyMsg:
		if cmd, handled := f.handleNavigationKey(msg); handled {
			return cmd
		}
	}

	c := f.focus.Current()
	if c == nil {
		return nil
	}

	_, cmd := c.Update(msg)

	return cmd
}

// buildRows renders every entry via its own View(), in order.
func (f *Model) buildRows() []string {
	out := make([]string, 0, len(f.entries))

	for _, m := range f.entries {
		out = append(out, m.View().Content)
	}

	return out
}

// RenderHint returns the focused entry's hint rendered with the form's hint
// styles, or "" when there is no hint. Callers render this below the form
// viewport so it stays visible regardless of scroll position.
func (f *Model) RenderHint() string {
	if c := f.focus.Current(); c != nil {
		if h, ok := c.(field.Hinted); ok {
			if hint := h.Hint(); hint != "" {
				return f.styles.HintBlock.Render(f.styles.HintText.Render(hint))
			}
		}
	}

	return ""
}

// refreshContent pushes the freshly-rendered rows into the viewport, so
// scroll adjustments (see syncScroll) are clamped against current content
// rather than whatever was last set.
func (f *Model) refreshContent() {
	f.vp.SetContent(lipgloss.JoinVertical(lipgloss.Left, f.buildRows()...))
}

// View renders every entry into the viewport.
func (f *Model) View() string {
	f.refreshContent()

	return f.vp.View()
}

// Keys returns the form's navigation bindings plus any bindings from the focused row.
func (f *Model) Keys() []key.Binding {
	bindings := []key.Binding{keyUp, keyDown}

	if c := f.focus.Current(); c != nil {
		bindings = append(bindings, c.Keys()...)
	}

	return bindings
}

// Validate calls each validatable entry's validator, sets its error, and
// returns true when all entries pass.
func (f *Model) Validate() bool {
	valid := true

	for _, e := range f.validatable {
		e.SetErr(nil)

		if msg := e.Validate(); msg != "" {
			e.SetErr(errors.New(msg))

			valid = false
		}
	}

	return valid
}

// FocusFirstError blurs the current row and focuses the first focusable
// entry whose Err is non-empty. Returns nil when there are no errors.
func (f *Model) FocusFirstError() tea.Cmd {
	for i, c := range f.focus.Items() {
		v, ok := c.(field.Validateable)
		if !ok || v.Err() == nil {
			continue
		}

		return f.focus.Set(i)
	}

	return nil
}

// focusedRowIndex returns the row index in buildRows for the focused row, or -1.
func (f *Model) focusedRowIndex() int {
	i := f.focus.Position()
	if i < 0 {
		return -1
	}

	return f.focusableLine[i]
}

// rowRange returns the [start, end] line range (inclusive) that buildRows()'s
// i'th row occupies within the joined content.
func (f *Model) rowRange(i int) (start, end int) {
	rows := f.buildRows()

	for _, r := range rows[:i] {
		start += lipgloss.Height(r)
	}

	end = start + lipgloss.Height(rows[i]) - 1

	return start, end
}

// ensureVisible scrolls the viewport by the minimum amount to bring [start, end] into view.
func (f *Model) ensureVisible(start, end int) {
	height := f.vp.Height()

	switch {
	case start < f.vp.YOffset():
		f.vp.SetYOffset(start)
	case end >= f.vp.YOffset()+height:
		f.vp.SetYOffset(end - height + 1)
	}
}

// syncScroll scrolls the viewport to keep the focused row visible. For
// field.CursorAware fields whose extent exceeds the viewport, narrows to the cursor line.
func (f *Model) syncScroll() {
	i := f.focusedRowIndex()
	if i < 0 {
		return
	}

	start, end := f.rowRange(i)

	if s, ok := f.focus.Current().(Sizeable); ok {
		if ca, ok := s.Unwrap().(field.CursorAware); ok && end-start+1 > f.vp.Height() {
			end = start + ca.CursorLine()
		}
	}

	f.ensureVisible(start, end)
}

// focusedEntryClaimsKey reports whether the currently focused entry claims k
// via its own Keys() bindings - e.g. a closed selectfield claims Enter to
// open its picker, or an open one claims ↑/↓ for dropdown navigation.
func (f *Model) focusedEntryClaimsKey(k tea.KeyMsg) bool {
	c := f.focus.Current()
	if c == nil {
		return false
	}

	for _, b := range c.Keys() {
		if key.Matches(k, b) {
			return true
		}
	}

	return false
}

// focusedIsField reports whether the currently focused entry is a Sizeable
// (field) entry, as opposed to a control.
func (f *Model) focusedIsField() bool {
	c := f.focus.Current()
	if c == nil {
		return false
	}

	_, ok := c.(Sizeable)

	return ok
}

// handleNavigationKey intercepts ↑/↓/↵ for the form's own row-to-row
// navigation. For ↑/↓, when the focused entry manages its own nested focus
// (field.FocusModel, e.g. a buttonstack.Model), the key is forwarded to it
// first; the form only takes over once the child reports it has nothing
// left to move to internally (its Position() didn't change). Other entries
// keep the existing claims-based prediction, since "did it handle it" has
// to be known before forwarding for them.
// Returns (cmd, true) when it handled the key, (nil, false) to let it fall through.
func (f *Model) handleNavigationKey(k tea.KeyMsg) (cmd tea.Cmd, handled bool) {
	if key.Matches(k, keyEnter) && f.focusedIsField() && !f.focusedEntryClaimsKey(k) {
		return f.focus.Shift(1), true
	}

	var dir int

	switch {
	case key.Matches(k, keyUp):
		dir = -1
	case key.Matches(k, keyDown):
		dir = 1
	default:
		return nil, false
	}

	c := f.focus.Current()
	if c == nil {
		return nil, false
	}

	if n, ok := c.(field.FocusModel); ok {
		before := n.Position()
		_, cmd := c.Update(k)

		if n.Position() != before {
			return cmd, true
		}

		return tea.Batch(cmd, f.focus.Shift(dir)), true
	}

	if !f.focusedEntryClaimsKey(k) {
		return f.focus.Shift(dir), true
	}

	return nil, false
}
