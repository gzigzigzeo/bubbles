package scrollview_test

import (
	"strings"
	"testing"

	"github.com/gzigzigzeo/bubbles/scrollview"
	"github.com/stretchr/testify/require"
)

func newViewForTest(height int, lines int) *scrollview.Model {
	model := scrollview.New()
	model.SetHeight(height)
	model.SetWidth(10)
	model.SetContent(strings.Repeat("x\n", lines-1) + "x")

	return model
}

func TestScrollTo_NoScrollWhenVisible(t *testing.T) {
	model := newViewForTest(5, 20)
	model.SetYOffset(5)

	model.ScrollTo(6)

	require.Equal(t, 5, model.YOffset())
}

func TestScrollTo_SmoothScrollUp(t *testing.T) {
	model := newViewForTest(5, 20)
	model.SetYOffset(10)

	// Target is one line above the viewport; should scroll up by exactly one.
	model.ScrollTo(9)

	require.Equal(t, 9, model.YOffset())
}

func TestScrollTo_SmoothScrollUpFar(t *testing.T) {
	model := newViewForTest(5, 20)
	model.SetYOffset(10)

	// Target is several lines above the viewport; scroll up by one line
	// (the next render will call ScrollTo again if focus keeps moving).
	model.ScrollTo(5)

	require.Equal(t, 9, model.YOffset())
}

func TestScrollTo_Exception1_MaximiseContextAbove(t *testing.T) {
	model := newViewForTest(5, 20)
	model.SetYOffset(10)

	// Target is near the top of the content and above the viewport.
	// The viewport should scroll to the top, placing the target at the bottom
	// so rows above it are visible.
	model.ScrollTo(3)

	require.Equal(t, 0, model.YOffset())
}

func TestScrollTo_Exception1_TargetAtBottomOfViewport(t *testing.T) {
	model := newViewForTest(5, 20)
	model.SetYOffset(0)

	// Target is within the first viewport-height lines and currently at the
	// bottom of the visible area. Maximising context above keeps it visible.
	model.ScrollTo(4)

	require.Equal(t, 0, model.YOffset())
}

func TestScrollTo_SmoothScrollDown(t *testing.T) {
	model := newViewForTest(5, 20)
	model.SetYOffset(5)

	// Target is one line below the viewport; should scroll down by exactly one.
	model.ScrollTo(10)

	require.Equal(t, 6, model.YOffset())
}

func TestScrollTo_Exception2_MaximiseContextBelow(t *testing.T) {
	model := newViewForTest(5, 20)
	model.SetYOffset(5)

	// Target is near the bottom of the content and below the viewport.
	// The viewport should scroll to the bottom, placing the target at the top
	// so rows below it are visible.
	model.ScrollTo(18)

	require.Equal(t, 15, model.YOffset()) // total(20) - height(5)
}

func TestScrollTo_Exception2_TargetAtTopOfViewport(t *testing.T) {
	model := newViewForTest(5, 20)
	model.SetYOffset(15)

	// Target is within the last viewport-height lines and currently at the
	// top of the visible area. Maximising context below keeps it visible.
	model.ScrollTo(15)

	require.Equal(t, 15, model.YOffset())
}

func TestScrollTo_ClampTop(t *testing.T) {
	model := newViewForTest(5, 20)
	model.SetYOffset(10)

	model.ScrollTo(-5)

	require.Equal(t, 0, model.YOffset())
}

func TestScrollTo_ClampBottom(t *testing.T) {
	model := newViewForTest(5, 20)
	model.SetYOffset(0)

	model.ScrollTo(100)

	require.Equal(t, 15, model.YOffset()) // total(20) - height(5)
}

func TestScrollTo_SmallContent(t *testing.T) {
	model := newViewForTest(10, 5)
	model.SetYOffset(0)

	model.ScrollTo(3)

	require.Equal(t, 0, model.YOffset())
}
