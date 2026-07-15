// Package scrollview extends [charm.land/bubbles/v2/viewport.Model] with a
// 1-column scrollbar. The scrollbar is only visible when content overflows
// the viewport height; when content fits, the column is rendered blank.
// The scrollbar is fully styleable via [Styles].
package scrollview

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ScrollbarPosition is the side on which the scrollbar column is rendered.
type ScrollbarPosition int

const (
	// Left places the scrollbar column to the left of the content.
	Left ScrollbarPosition = iota
	// Right places the scrollbar column to the right of the content.
	Right
)

const reservedEndCells = 2

// Model extends [charm.land/bubbles/v2/viewport.Model] with a 1-column
// scrollbar alongside the content.
//
// All standard viewport methods (SetHeight, Height, YOffset, SetYOffset,
// Init, etc.) are promoted directly from the embedded viewport.Model.
// SetWidth and Width are overridden to account for the scrollbar column.
// SetContent is overridden to track the total line count for thumb placement.
//
// The zero value is not usable; create instances with [New].
type Model struct {
	viewport.Model

	// Position controls which side the scrollbar column appears on.
	Position ScrollbarPosition

	// Styles holds the scrollbar styles.
	Styles Styles

	total int // total number of content lines
}

// New returns a pointer to a Model with default styles.
func New() *Model {
	return &Model{
		Model:    viewport.New(),
		Position: Left,
		Styles:   DefaultStyles(),
		total:    0,
	}
}

// Update forwards messages to the embedded viewport and returns an updated copy.
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	vp, cmd := m.Model.Update(msg)
	m.Model = vp

	return m, cmd
}

// SetWidth sets the total component width. The embedded viewport receives
// width-1 to reserve a column for the scrollbar.
func (m *Model) SetWidth(w int) {
	inner := max(w-lipgloss.Width(m.Styles.Track.Render()), 0)

	m.Model.SetWidth(inner)
}

// Width returns the total component width including the scrollbar column.
func (m *Model) Width() int {
	return m.Model.Width() + lipgloss.Width(m.Styles.Track.Render())
}

// SetStyles replaces the scrollbar styles.
func (m *Model) SetStyles(s Styles) {
	m.Styles = s
}

// SetContent sets the viewport content. The line count is tracked for
// scrollbar thumb placement.
func (m *Model) SetContent(s string) {
	m.total = lipgloss.Height(s)
	m.Model.SetContent(s)
}

// ScrollTo scrolls the viewport so that line is within the visible region.
// If line is already visible no change is made.
//
// By default the viewport scrolls the minimum amount: one line up when the
// target moves above the visible area, or just enough down to bring the target
// into view. At the content boundaries it maximises context around the target
// row: when the target is near the top of the content the viewport is
// positioned so the target sits at the bottom, revealing any rows above it;
// when the target is near the bottom the viewport is positioned so the target
// sits at the top, revealing any rows below it.
func (m *Model) ScrollTo(line int) {
	height := m.Height()
	total := m.total

	if line < m.YOffset() {
		offset := m.YOffset() - 1

		// Near the top of the content: maximise visible rows above the target
		// while keeping it in view (e.g. a heading above the first focusable
		// control).
		if line <= height-1 {
			offset = max(line-(height-1), 0)
		}

		m.SetYOffset(offset)

		return
	}

	if line >= m.YOffset()+height {
		offset := line - height + 1

		// Near the bottom of the content: maximise visible rows below the
		// target while keeping it in view (e.g. trailing non-focusable rows
		// after the last control).
		if line >= total-height {
			offset = min(line, total-height)
		}

		m.SetYOffset(offset)
	}
}

// View renders the embedded viewport alongside its scrollbar column.
func (m *Model) View() string {
	raw := m.Model.View()
	if raw == "" {
		return ""
	}

	scrollbar := m.scrollbarColumn()
	if scrollbar == "" {
		return raw
	}

	if m.Position == Left {
		return lipgloss.JoinHorizontal(lipgloss.Top, scrollbar, raw)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, raw, scrollbar)
}

// scrollbarColumn returns the 1-column scrollbar as a newline-joined string
// of height styled characters.
func (m *Model) scrollbarColumn() string {
	height := m.Height()
	if height == 0 {
		return ""
	}

	cells := make([]string, height)

	if m.total <= height {
		return renderHiddenScrollbar(cells, m.Styles.HiddenBar)
	}

	return renderVisibleScrollbar(m, cells, height)
}

func renderHiddenScrollbar(cells []string, style lipgloss.Style) string {
	for i := range cells {
		cells[i] = style.Render()
	}

	return lipgloss.JoinVertical(lipgloss.Top, cells...)
}

func renderVisibleScrollbar(model *Model, cells []string, height int) string {
	// The thumb occupies exactly one cell in the inner range [1, h-2] so
	// cells 0 and h-1 are always reserved for the top/bottom indicators.
	innerCells := height - reservedEndCells
	thumbPos := 1

	maxOffset := model.total - height

	if innerCells > 0 && maxOffset > 0 {
		thumbPos = 1 + model.YOffset()*(innerCells-1)/maxOffset
	}

	for i := range cells {
		if i == thumbPos && innerCells > 0 {
			cells[i] = model.Styles.Thumb.Render()
		} else {
			cells[i] = model.Styles.Track.Render()
		}
	}

	cells[0] = topIndicator(model)
	cells[height-1] = bottomIndicator(model, height)

	return lipgloss.JoinVertical(lipgloss.Top, cells...)
}

func topIndicator(model *Model) string {
	if model.YOffset() > 0 {
		return model.Styles.MoreAbove.Render()
	}

	return model.Styles.NoMoreAbove.Render()
}

func bottomIndicator(model *Model, height int) string {
	if model.YOffset()+height < model.total {
		return model.Styles.MoreBelow.Render()
	}

	return model.Styles.NoMoreBelow.Render()
}
