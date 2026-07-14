package navigator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoordinator_FocusFollowScroll(t *testing.T) {
	// Content:
	//   0: 0
	//   1: 1
	//   2: 2
	//   3: 3
	//   4: 4
	// Height 3.
	nav := New(
		&testItem{text: "0"},
		&testItem{text: "1"},
		&testItem{text: "2"},
		&testItem{text: "3"},
		&testItem{text: "4"},
	)
	nav.ViewportController().SetHeight(3)
	_ = nav.FocusFirst()

	require.Equal(t, 0, nav.ViewportController().YOffset())

	// Move down to item 3 (line 3). It is below the viewport, so the viewport
	// scrolls by the minimum amount.
	sendKey(t, nav, "down")
	sendKey(t, nav, "down")
	sendKey(t, nav, "down")

	require.Equal(t, 3, nav.ctrl.FocusedIndex())
	require.Equal(t, 1, nav.ViewportController().YOffset())

	got := nav.CursorLine()
	require.GreaterOrEqual(t, got, nav.ViewportController().YOffset())
	require.Less(t, got, nav.ViewportController().YOffset()+nav.ViewportController().Height())
}

func TestCoordinator_BoundaryScrollUp(t *testing.T) {
	// Content:
	//   0: header
	//   1: Alpha
	//   2: Beta
	// Height 2.
	nav := New(
		testLabel("header"),
		&testItem{text: "Alpha"},
		&testItem{text: "Beta"},
	)
	nav.ViewportController().SetHeight(2)
	_ = nav.FocusFirst()

	// FocusFirst with minimum scroll: Alpha at line 1 is visible (viewport 0-1).
	require.Equal(t, 0, nav.ViewportController().YOffset())

	// Move down to Beta (line 2). Viewport scrolls so Beta is at bottom (1-2).
	sendKey(t, nav, "down")
	require.Equal(t, 1, nav.ViewportController().YOffset())

	// Up at Alpha: focus moves to Alpha (line 1), which is visible, no scroll.
	sendKey(t, nav, "up")
	require.Equal(t, 1, nav.ViewportController().YOffset())

	// Up again at Alpha. No focusable above, so boundary scroll reveals header.
	sendKey(t, nav, "up")
	require.Equal(t, 0, nav.ViewportController().YOffset())
}

func TestCoordinator_BoundaryScrollDown(t *testing.T) {
	// Content:
	//   0: header0
	//   1: header1
	//   2: Beta (last focusable)
	//   3: footer0
	//   4: footer1
	//   5: footer2
	// Height 3.
	nav := New(
		testLabel("header0"),
		testLabel("header1"),
		&testItem{text: "Beta"},
		testLabel("footer0"),
		testLabel("footer1"),
		testLabel("footer2"),
	)
	nav.ViewportController().SetHeight(3)
	_ = nav.FocusFirst()

	require.Equal(t, 0, nav.ViewportController().YOffset())

	// Down at Beta (last focusable). Boundary scroll down one line.
	sendKey(t, nav, "down")
	require.Equal(t, 1, nav.ViewportController().YOffset())

	sendKey(t, nav, "down")
	require.Equal(t, 2, nav.ViewportController().YOffset())

	// Beta is now at the top; no more content below, so stop.
	sendKey(t, nav, "down")
	require.Equal(t, 2, nav.ViewportController().YOffset())
}

func TestCoordinator_StopsAtScreenEdge(t *testing.T) {
	// Content:
	//   0: header1
	//   1: header2
	//   2: Alpha
	//   3: Beta
	//   4: footer
	// Height 3.
	nav := New(
		testLabel("header1"),
		testLabel("header2"),
		&testItem{text: "Alpha"},
		&testItem{text: "Beta"},
		testLabel("footer"),
	)
	nav.ViewportController().SetHeight(3)
	_ = nav.FocusFirst()
	nav.ViewportController().SetYOffset(2)

	require.Equal(t, 2, nav.ViewportController().YOffset())

	// Up at Alpha boundary scrolls until Alpha reaches bottom.
	sendKey(t, nav, "up")
	require.Equal(t, 1, nav.ViewportController().YOffset())

	sendKey(t, nav, "up")
	require.Equal(t, 0, nav.ViewportController().YOffset())

	// No more content above.
	sendKey(t, nav, "up")
	require.Equal(t, 0, nav.ViewportController().YOffset())
}

func TestCoordinator_NestedNavigatorFollowsViewport(t *testing.T) {
	inner := New(
		&testItem{text: "Echo"},
		&testItem{text: "Foxtrot"},
		&testItem{text: "Golf"},
		&testItem{text: "Hotel"},
		&testItem{text: "India"},
	)

	// Outer content:
	//   0: header
	//   1: Alpha
	//   2: Echo
	//   3: Foxtrot
	//   4: Golf
	//   5: Hotel
	//   6: India
	// Height 3.
	nav := New(
		testLabel("header"),
		&testItem{text: "Alpha"},
		inner,
	)
	nav.ViewportController().SetHeight(3)
	_ = nav.FocusFirst()
	sendKey(t, nav, "down") // move into inner -> Echo

	require.Equal(t, 0, inner.ctrl.FocusedIndex())

	for _, wantCursor := range []int{3, 4, 5, 6} {
		sendKey(t, nav, "down")

		got := nav.CursorLine()
		require.Equal(t, wantCursor, got)
		require.GreaterOrEqual(t, got, nav.ViewportController().YOffset())
		require.Less(t, got, nav.ViewportController().YOffset()+nav.ViewportController().Height())
	}

	for _, wantCursor := range []int{5, 4, 3, 2} {
		sendKey(t, nav, "up")

		got := nav.CursorLine()
		require.Equal(t, wantCursor, got)
		require.GreaterOrEqual(t, got, nav.ViewportController().YOffset())
		require.Less(t, got, nav.ViewportController().YOffset()+nav.ViewportController().Height())
	}
}
