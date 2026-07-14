package scrollview

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func newViewForTest(h int, lines int) Model {
	m := New()
	m.SetHeight(h)
	m.SetWidth(10)
	m.SetContent(strings.Repeat("x\n", lines-1) + "x")

	return m
}

func TestScrollTo_NoScrollWhenVisible(t *testing.T) {
	m := newViewForTest(5, 20)
	m.SetYOffset(5)

	m.ScrollTo(6)

	require.Equal(t, 5, m.YOffset())
}

func TestScrollTo_SmoothScrollUp(t *testing.T) {
	m := newViewForTest(5, 20)
	m.SetYOffset(10)

	// Target is one line above the viewport; should scroll up by exactly one.
	m.ScrollTo(9)

	require.Equal(t, 9, m.YOffset())
}

func TestScrollTo_SmoothScrollUpFar(t *testing.T) {
	m := newViewForTest(5, 20)
	m.SetYOffset(10)

	// Target is several lines above the viewport; scroll up by one line
	// (the next render will call ScrollTo again if focus keeps moving).
	m.ScrollTo(5)

	require.Equal(t, 9, m.YOffset())
}

func TestScrollTo_Exception1_MaximiseContextAbove(t *testing.T) {
	m := newViewForTest(5, 20)
	m.SetYOffset(10)

	// Target is near the top of the content and above the viewport.
	// The viewport should scroll to the top, placing the target at the bottom
	// so rows above it are visible.
	m.ScrollTo(3)

	require.Equal(t, 0, m.YOffset())
}

func TestScrollTo_Exception1_TargetAtBottomOfViewport(t *testing.T) {
	m := newViewForTest(5, 20)
	m.SetYOffset(0)

	// Target is within the first viewport-height lines and currently at the
	// bottom of the visible area. Maximising context above keeps it visible.
	m.ScrollTo(4)

	require.Equal(t, 0, m.YOffset())
}

func TestScrollTo_SmoothScrollDown(t *testing.T) {
	m := newViewForTest(5, 20)
	m.SetYOffset(5)

	// Target is one line below the viewport; should scroll down by exactly one.
	m.ScrollTo(10)

	require.Equal(t, 6, m.YOffset())
}

func TestScrollTo_Exception2_MaximiseContextBelow(t *testing.T) {
	m := newViewForTest(5, 20)
	m.SetYOffset(5)

	// Target is near the bottom of the content and below the viewport.
	// The viewport should scroll to the bottom, placing the target at the top
	// so rows below it are visible.
	m.ScrollTo(18)

	require.Equal(t, 15, m.YOffset()) // total(20) - h(5)
}

func TestScrollTo_Exception2_TargetAtTopOfViewport(t *testing.T) {
	m := newViewForTest(5, 20)
	m.SetYOffset(15)

	// Target is within the last viewport-height lines and currently at the
	// top of the visible area. Maximising context below keeps it visible.
	m.ScrollTo(15)

	require.Equal(t, 15, m.YOffset())
}

func TestScrollTo_ClampTop(t *testing.T) {
	m := newViewForTest(5, 20)
	m.SetYOffset(10)

	m.ScrollTo(-5)

	require.Equal(t, 0, m.YOffset())
}

func TestScrollTo_ClampBottom(t *testing.T) {
	m := newViewForTest(5, 20)
	m.SetYOffset(0)

	m.ScrollTo(100)

	require.Equal(t, 15, m.YOffset()) // total(20) - h(5)
}

func TestScrollTo_SmallContent(t *testing.T) {
	m := newViewForTest(10, 5)
	m.SetYOffset(0)

	m.ScrollTo(3)

	require.Equal(t, 0, m.YOffset())
}
