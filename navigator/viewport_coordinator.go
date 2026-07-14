// Package navigator manages an ordered list of rows with keyboard focus.
package navigator

import (
	tea "charm.land/bubbletea/v2"
)

// Viewport is the minimal surface a [ViewportCoordinator] needs to render a
// clipped view. The standard use case is a pointer to a scrollview Model.
type Viewport interface {
	SetContent(string)
	SetYOffset(int)
	SetHeight(int)
	View() string
}

// ViewportCoordinator sits between a [Model] and an external viewport. It tracks
// the viewport height and scroll offset and adjusts the offset so the focused
// row stays visible, including one-line-at-a-time scrolling at the first/last
// focusable row to reveal non-focusable content beyond the boundary.
type ViewportCoordinator struct {
	height   int
	yOffset  int
	viewport Viewport
}

// NewViewportCoordinator creates a ViewportCoordinator with zero height. Set the
// height with [SetHeight] before use, and optionally attach a viewport with
// [SetViewport].
func NewViewportCoordinator() *ViewportCoordinator {
	return &ViewportCoordinator{}
}

// SetHeight sets the viewport height in lines. When a viewport is attached its
// height is kept in sync.
func (c *ViewportCoordinator) SetHeight(h int) {
	c.height = max(h, 0)

	if c.viewport != nil {
		c.viewport.SetHeight(c.height)
	}
}

// Height returns the configured viewport height.
func (c *ViewportCoordinator) Height() int {
	return c.height
}

// YOffset returns the current top line of the visible viewport. The parent
// should apply this offset to the paired viewport (e.g. via SetYOffset), or it
// can attach the viewport with [SetViewport] and let the coordinator render it.
func (c *ViewportCoordinator) YOffset() int {
	return c.yOffset
}

// SetYOffset sets the top line of the visible viewport. The offset is clamped
// to be non-negative; it is further constrained to the content bounds on the
// next scroll operation.
func (c *ViewportCoordinator) SetYOffset(y int) {
	c.yOffset = max(y, 0)

	if c.viewport != nil {
		c.viewport.SetYOffset(c.yOffset)
	}
}

// SetViewport attaches a viewport. Once attached, height and offset changes are
// forwarded automatically and [View] renders the clipped viewport content.
func (c *ViewportCoordinator) SetViewport(v Viewport) {
	c.viewport = v

	if v != nil {
		v.SetHeight(c.height)
		v.SetYOffset(c.yOffset)
	}
}

// View renders the attached viewport with the current content offset. If no
// viewport is attached it returns an empty view.
func (c *ViewportCoordinator) View() tea.View {
	if c.viewport == nil {
		return tea.View{}
	}

	return tea.NewView(c.viewport.View())
}

// SetContent updates the attached viewport content. It is a no-op when no
// viewport is attached.
func (c *ViewportCoordinator) SetContent(content string) {
	if c.viewport != nil {
		c.viewport.SetContent(content)
	}
}

// ScrollToFocus adjusts yOffset so the current cursor line is visible. It
// scrolls the minimum amount: the cursor lands at the top edge when it moves
// above the viewport and at the bottom edge when it moves below.
func (c *ViewportCoordinator) ScrollToFocus(nav *Model) {
	c.scrollToFocus(nav.CursorLine(), nav.totalLines())
}

// scrollToFocus is the internal implementation of ScrollToFocus.
func (c *ViewportCoordinator) scrollToFocus(cursor, total int) {
	if c.height <= 0 {
		return
	}

	if total <= c.height {
		c.yOffset = 0
		c.syncYOffset()

		return
	}

	if cursor < c.yOffset {
		c.yOffset = cursor
		c.syncYOffset()

		return
	}

	if cursor >= c.yOffset+c.height {
		c.yOffset = cursor - c.height + 1
	}

	c.clampYOffset(total)
	c.syncYOffset()
}

// scrollAtBoundary scrolls the viewport one line in dir when focus is already
// at the boundary and cannot move. In open mode this reveals non-focusable
// content above/below the boundary while keeping the focused row visible.
func (c *ViewportCoordinator) scrollAtBoundary(cursor, total, dir int) {
	if c.height <= 0 {
		return
	}

	rel := cursor - c.yOffset

	if dir < 0 { // up
		if c.yOffset > 0 && rel < c.height-1 {
			c.yOffset--
		}
	} else { // down
		if c.yOffset+c.height < total && rel > 0 {
			c.yOffset++
		}
	}

	c.clampYOffset(total)
	c.syncYOffset()
}

// clampYOffset keeps yOffset within the current content bounds for the given
// total line count. A negative total leaves yOffset unchanged.
func (c *ViewportCoordinator) clampYOffset(total int) {
	if total < 0 {
		return
	}

	if total <= c.height {
		c.yOffset = 0

		return
	}

	c.yOffset = max(0, min(c.yOffset, total-c.height))
}

// syncYOffset forwards the current offset to the attached viewport.
func (c *ViewportCoordinator) syncYOffset() {
	if c.viewport != nil {
		c.viewport.SetYOffset(c.yOffset)
	}
}
