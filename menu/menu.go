// Package menu provides a reusable cursor-driven list of choices, scrolled
// by an embedded viewport.Model, for screens that need to present the user
// with a list of named (and optionally described) options.
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

// ChoiceMsg is the message a Menu fires when the user selects an option. Option
// points at the Option itself, for callers that also want its
// Name/Description alongside Value.
type ChoiceMsg[T any] struct {
	Value  T
	Option *Option[T]
}

// Model is a cursor-driven, optionally height-limited list of Options,
// scrolled by an embedded viewport.Model. It exposes the key bindings it
// responds to via Keys, but never renders a hint bar itself — callers own
// that. T is constrained to comparable so the cursor can be positioned by
// value (SetCursor/Cursor) instead of by index.
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

// SetWidth sets the width of the menu's viewport and the width each row is
// truncated to. w <= 0 resets to unlimited: rows are rendered at their
// natural length, uncapped.
func (m *Model[T]) SetWidth(w int) {
	m.width = w
	m.vp.SetWidth(w)
	m.syncContent()
	m.applyHeight()
	m.scrollCursorIntoView()
}

// SetHeight limits how many option rows View() renders, scrolling
// internally past that. h <= 0 resets to unlimited (every option shown, no
// scrolling).
func (m *Model[T]) SetHeight(h int) {
	m.height = h
	m.applyHeight()
	m.scrollCursorIntoView()
}

// applyHeight resolves the configured height onto the viewport. h <= 0
// means unlimited: every option is shown, recomputed fresh off the current
// option count so it stays correct after SetOptions changes it.
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

// CursorLine returns the 0-indexed line within View()'s rendered output that
// the cursor is currently on, after the menu's own internal scrolling has
// been applied. A caller embedding the menu inside a taller scrollable
// container (e.g. Form) can use this to keep the cursor's actual line
// visible as the user navigates, rather than the menu's full extent.
func (m *Model[T]) CursorLine() int {
	return m.cursor - m.vp.YOffset()
}

// SetOptions replaces the menu's options, e.g. when a dependent field (such
// as a chosen region or architecture) changes what should be selectable. The
// cursor resets to the first option; the marker is left untouched, so
// callers whose committed value no longer appears in the new list should
// call SetMarker (or SetCursor) with a valid value afterwards.
func (m *Model[T]) SetOptions(options []Option[T]) {
	m.options = options
	m.cursor = 0
	m.syncNameWidth()
	m.syncContent()
	m.applyHeight()
	m.scrollCursorIntoView()
}

// SetMarker marks the option whose Value equals value with
// Styles.CursorMarked/LabelMarked instead of the normal blurred styling,
// independent of the cursor — e.g. to show which value is currently
// committed while the user browses elsewhere. Multiple markers aren't
// supported; call it again to move the mark.
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

// scrollCursorIntoView scrolls the viewport by the minimum amount needed to
// bring the cursor's row into view. Scrolling up reveals it at the window's
// top (matching viewport.EnsureVisible's own behavior); scrolling down
// reveals it at the window's bottom instead of making it the new top —
// viewport.EnsureVisible always does the latter, which reads as a full-page
// jump instead of a smooth one-row scroll.
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

// Keys returns the key bindings this menu responds to, for callers to
// include in their own hint bar. It deliberately omits the viewport's bonus
// PgUp/PgDn/mouse-wheel scrolling, which is a free enhancement rather than a
// promised interaction.
func (m *Model[T]) Keys() []key.Binding {
	return []key.Binding{menuKeyUp, menuKeyDown, menuKeySelect}
}

// Init satisfies the tea.Model interface.
func (m *Model[T]) Init() tea.Cmd {
	return m.vp.Init()
}

// Update moves the cursor on Up/Down (no wraparound) and fires a Choice[T]
// for the selected option on Select. Every other message — including the
// viewport's own PgUp/PgDn/mouse-wheel handling — is forwarded to the
// embedded viewport.
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

// fireChoice returns a Cmd emitting a Choice[T] for the currently selected
// option. The Option pointer is into m.options, which is never resized
// after construction, so it stays valid for the Menu's lifetime.
func (m *Model[T]) fireChoice() tea.Cmd {
	opt := &m.options[m.cursor]
	return func() tea.Msg { return ChoiceMsg[T]{Value: opt.Value, Option: opt} }
}

// View renders the menu's viewport, its leftmost column carrying the scroll
// and cursor indicators (see gutter).
func (m *Model[T]) View() tea.View {
	return tea.NewView(m.vp.View())
}

// gutter renders the leftmost column: the scroll indicator (ScrollUp on the
// viewport's first visible line when there's content above it, ScrollDown
// on the last visible line when there's content below, blank otherwise)
// immediately followed by the cursor glyph (CursorFocused/CursorBlurred).
// If the viewport is only one line tall, "scroll up" wins over "scroll
// down" — an edge case not worth a second indicator on one row.
func (m *Model[T]) gutter(ctx viewport.GutterContext) string {
	return m.scrollIndicator(ctx) + m.cursorIndicator(ctx.Index)
}

// scrollIndicator returns ScrollUp/ScrollDown for the viewport's first/last
// visible line when there's more content in that direction, or a same-width
// blank otherwise. Uses Height() rather than VisibleLineCount() — the latter
// calls back into maxWidth(), which itself probes LeftGutterFunc to measure
// its width, causing infinite recursion. Height() is the configured height,
// which equals the actual visible line count whenever there's enough
// content to scroll (the only case the "more below" arrow can ever apply).
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

// cursorIndicator returns the cursor glyph for the currently-selected
// option's line, CursorMarked if it's the marked option's line (and not the
// cursor's), CursorBlurred otherwise, or a same-width blank if line is out
// of range.
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

// syncContent rebuilds every row, reflecting the current cursor position via
// label color (the cursor glyph itself is rendered by gutter), and pushes
// them into the viewport. Description is rendered at its natural length,
// never wrapped; rows wider than the configured width are truncated with a
// trailing ellipsis instead (see availableRowWidth).
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
// (scroll indicator immediately followed by the cursor glyph), matching
// what gutter always produces regardless of which branch scrollIndicator/
// cursorIndicator take on a given line.
func (m *Model[T]) gutterWidth() int {
	return lipgloss.Width(m.styles.ScrollUp.Render()) + lipgloss.Width(m.styles.CursorFocused.Render())
}
