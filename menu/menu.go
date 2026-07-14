// Package menu provides a cursor-driven list of named choices
// scrolled by an embedded viewport.Model.
package menu

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var menuKeyUp = key.NewBinding(
	key.WithKeys("up", "k"),
	key.WithHelp("↑/k", "up"),
)

var menuKeyDown = key.NewBinding(
	key.WithKeys("down", "j"),
	key.WithHelp("↓/j", "down"),
)

var menuKeySelect = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "select"),
)

const (
	defaultWidth = 80
)

// Option is a single row in a Menu, carrying the value emitted when it's chosen.
type Option[T any] struct {
	Name        string
	Description string // optional; if no option sets one, the description column is omitted
	Value       T
}

// ChoiceMsg is fired when the user selects an option. Option
// points at the Option alongside its Value.
type ChoiceMsg[T any] struct {
	Value  T
	Option *Option[T]
}

// Model is a cursor-driven, optionally height-limited list of Options
// scrolled by an embedded viewport.Model.
type Model[T comparable] struct {
	options         []Option[T]
	widestNameWidth int // cached widestName(), recomputed only when options change
	styles          Styles
	cursor          int
	height          int // 0 means unlimited: every line is shown
	width           int // 0 means unlimited: rows are never truncated
	marker          T
	hasMarker       bool

	vp viewport.Model
}

// New creates a Menu over options. By default every option is shown with no
// scrolling; call SetHeight to limit the visible rows.
func New[T comparable](options []Option[T]) *Model[T] {
	m := &Model[T]{
		options: options,
		vp:      viewport.New(),
	}
	m.vp.LeftGutterFunc = m.gutter

	m.syncNameWidth()
	m.SetWidth(defaultWidth)

	return m
}

// SetStyles sets the visual styles used to render the menu.
func (m *Model[T]) SetStyles(s Styles) {
	m.styles = s
	m.syncContent()
	m.applyHeight()
	m.scrollCursorIntoView()
}

// SetWidth sets the viewport and row truncation width.
// w <= 0 resets to unlimited (rows rendered at natural length).
func (m *Model[T]) SetWidth(w int) {
	m.width = w
	m.vp.SetWidth(w)
	m.syncContent()
	m.applyHeight()
	m.scrollCursorIntoView()
}

// SetHeight limits how many option rows View() renders.
// h <= 0 resets to unlimited (every option shown, no scrolling).
func (m *Model[T]) SetHeight(h int) {
	m.height = h
	m.applyHeight()
	m.scrollCursorIntoView()
}

// applyHeight resolves the configured height onto the viewport.
// h <= 0 means unlimited, recomputed from the current option count.
func (m *Model[T]) applyHeight() {
	h := m.height
	if h <= 0 {
		h = len(m.options)
	}

	m.vp.SetHeight(max(1, h))
}

// Cursor returns the Value of the currently highlighted option.
func (m *Model[T]) Cursor() T {
	return m.options[m.cursor].Value
}

// CursorLine returns the 0-indexed line of the cursor within View(),
// after internal scrolling is applied.
func (m *Model[T]) CursorLine() int {
	return m.cursor - m.vp.YOffset()
}

// SetOptions replaces the menu's options and resets the cursor to zero.
// The marker is left untouched; call SetMarker/SetValue as needed.
func (m *Model[T]) SetOptions(options []Option[T]) {
	m.options = options
	m.cursor = 0
	m.syncNameWidth()
	m.syncContent()
	m.applyHeight()
	m.scrollCursorIntoView()
}

// SetMarker marks the option matching value with CursorMarked/LabelMarked,
// independent of the cursor. Call again to move the mark.
func (m *Model[T]) SetMarker(value T) {
	m.marker = value
	m.hasMarker = true
	m.syncContent()
}

// SetValue moves the cursor to the option whose Value equals value,
// scrolling its row fully into view. It's a no-op if no option matches.
func (m *Model[T]) SetValue(value T) {
	for i, opt := range m.options {
		if opt.Value == value {
			m.setCursorIndex(i)
			return
		}
	}
}

// setCursorIndex moves the cursor to i, clamped to the option range, and
// scrolls its row fully into view.
func (m *Model[T]) setCursorIndex(i int) {
	if len(m.options) == 0 {
		return
	}

	if i < 0 {
		i = 0
	}

	if i >= len(m.options) {
		i = len(m.options) - 1
	}

	m.cursor = i
	m.syncContent()
	m.scrollCursorIntoView()
}

// scrollCursorIntoView scrolls the viewport the minimum amount needed
// to bring the cursor's row into view.
func (m *Model[T]) scrollCursorIntoView() {
	line := m.cursor
	height := m.vp.Height()

	switch {
	case line < m.vp.YOffset():
		m.vp.SetYOffset(line)
	case line >= m.vp.YOffset()+height:
		m.vp.SetYOffset(line - height + 1)
	}
}

// Keys returns the key bindings this menu responds to,
// for callers to include in their own hint bar.
func (m *Model[T]) Keys() []key.Binding {
	return []key.Binding{menuKeyUp, menuKeyDown, menuKeySelect}
}

// Init satisfies the tea.Model interface.
func (m *Model[T]) Init() tea.Cmd {
	return m.vp.Init()
}

// Update moves the cursor on Up/Down and fires a ChoiceMsg on Select.
// All other messages are forwarded to the embedded viewport.
func (m *Model[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(km, menuKeyUp):
			m.setCursorIndex(m.cursor - 1)
			return m, nil
		case key.Matches(km, menuKeyDown):
			m.setCursorIndex(m.cursor + 1)
			return m, nil
		case key.Matches(km, menuKeySelect):
			return m, m.fireChoice()
		}
	}

	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)

	return m, cmd
}

// fireChoice returns a Cmd emitting a ChoiceMsg for the currently
// selected option.
func (m *Model[T]) fireChoice() tea.Cmd {
	opt := &m.options[m.cursor]
	return func() tea.Msg { return ChoiceMsg[T]{Value: opt.Value, Option: opt} }
}

// View renders the menu's viewport, its leftmost column carrying the scroll
// and cursor indicators (see gutter).
func (m *Model[T]) View() tea.View {
	return tea.NewView(m.vp.View())
}

// gutter renders the leftmost column: scroll indicator followed
// by the cursor glyph (CursorFocused/CursorBlurred).
func (m *Model[T]) gutter(ctx viewport.GutterContext) string {
	return m.scrollIndicator(ctx) + m.cursorIndicator(ctx.Index)
}

// scrollIndicator returns ScrollUp/ScrollDown on the first/last visible line
// when more content exists in that direction, or a same-width blank otherwise.
func (m *Model[T]) scrollIndicator(ctx viewport.GutterContext) string {
	last := m.vp.YOffset() + m.vp.Height() - 1

	switch {
	case ctx.Index == m.vp.YOffset() && m.vp.YOffset() > 0:
		return m.styles.ScrollUp.Render()
	case ctx.Index == last && last < ctx.TotalLines-1:
		return m.styles.ScrollDown.Render()
	default:
		return strings.Repeat(" ", lipgloss.Width(m.styles.ScrollUp.Render()))
	}
}

// cursorIndicator returns the cursor glyph for the given line:
// CursorFocused, CursorMarked, CursorBlurred, or a same-width blank.
func (m *Model[T]) cursorIndicator(line int) string {
	switch {
	case line < 0 || line >= len(m.options):
		return strings.Repeat(" ", lipgloss.Width(m.styles.CursorFocused.Render()))
	case line == m.cursor:
		return m.styles.CursorFocused.Render()
	case m.hasMarker && m.options[line].Value == m.marker:
		return m.styles.CursorMarked.Render()
	default:
		return m.styles.CursorBlurred.Render()
	}
}

// syncNameWidth recomputes the cached column width needed to align the
// Description column (0 if no option has a Description). Called only when
// m.options changes.
func (m *Model[T]) syncNameWidth() {
	widest := 0
	hasDesc := false

	for _, o := range m.options {
		if o.Description != "" {
			hasDesc = true
		}
		if len(o.Name) > widest {
			widest = len(o.Name)
		}
	}

	m.widestNameWidth = 0
	if hasDesc {
		m.widestNameWidth = widest + 5
	}
}

// syncContent rebuilds every row reflecting the current cursor position
// and pushes them into the viewport.
func (m *Model[T]) syncContent() {
	widest := m.widestNameWidth
	available := max(0, m.width-m.gutterWidth())

	lines := make([]string, len(m.options))

	for i, opt := range m.options {
		name := opt.Name
		if widest > 0 {
			name = lipgloss.PlaceHorizontal(widest, lipgloss.Left, name)
		}

		var label string
		switch {
		case i == m.cursor:
			label = m.styles.LabelFocused.Render(name)
		case m.hasMarker && opt.Value == m.marker:
			label = m.styles.LabelMarked.Render(name)
		default:
			label = m.styles.LabelBlurred.Render(name)
		}

		row := label
		if opt.Description != "" {
			row = lipgloss.JoinHorizontal(lipgloss.Top, label, m.styles.Description.Render(opt.Description))
		}

		if available > 0 {
			row = ansi.Truncate(row, available, "…")
		}

		lines[i] = row
	}

	m.vp.SetContent(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

// gutterWidth returns the rendered width of the leftmost gutter column
// (scroll indicator + cursor glyph).
func (m *Model[T]) gutterWidth() int {
	return lipgloss.Width(m.styles.ScrollUp.Render()) + lipgloss.Width(m.styles.CursorFocused.Render())
}
