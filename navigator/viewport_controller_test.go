package navigator_test

import (
	"testing"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/navigator"
	"github.com/gzigzigzeo/bubbles/navigator/rows/selectfield"
)

func TestCoordinator_FocusFollowScroll(t *testing.T) {
	// Content:
	//   0: 0
	//   1: 1
	//   2: 2
	//   3: 3
	//   4: 4
	// Height 3.
	nav := navigator.New(
		&testItem{text: "0", focused: false, disabled: false},
		&testItem{text: "1", focused: false, disabled: false},
		&testItem{text: "2", focused: false, disabled: false},
		&testItem{text: "3", focused: false, disabled: false},
		&testItem{text: "4", focused: false, disabled: false},
	)
	nav.ViewportController().SetHeight(3)
	_ = nav.FocusFirst()

	require.Equal(t, 0, nav.ViewportController().YOffset())

	// Move down to item 3 (line 3). It is below the viewport, so the viewport
	// scrolls by the minimum amount.
	sendKey(t, nav, "down")
	sendKey(t, nav, "down")
	sendKey(t, nav, "down")

	require.Equal(t, 3, nav.FocusedIndex())
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
	nav := navigator.New(
		testLabel("header"),
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: false},
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
	nav := navigator.New(
		testLabel("header0"),
		testLabel("header1"),
		&testItem{text: "Beta", focused: false, disabled: false},
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
	nav := navigator.New(
		testLabel("header1"),
		testLabel("header2"),
		&testItem{text: "Alpha", focused: false, disabled: false},
		&testItem{text: "Beta", focused: false, disabled: false},
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
	inner := navigator.New(
		&testItem{text: "Echo", focused: false, disabled: false},
		&testItem{text: "Foxtrot", focused: false, disabled: false},
		&testItem{text: "Golf", focused: false, disabled: false},
		&testItem{text: "Hotel", focused: false, disabled: false},
		&testItem{text: "India", focused: false, disabled: false},
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
	nav := navigator.New(
		testLabel("header"),
		&testItem{text: "Alpha", focused: false, disabled: false},
		inner,
	)
	nav.ViewportController().SetHeight(3)
	_ = nav.FocusFirst()
	sendKey(t, nav, "down") // move into inner -> Echo

	require.Equal(t, 0, inner.FocusedIndex())

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

func TestViewportController_View_ReappliesOffsetAfterContentChange(t *testing.T) {
	// Regression: viewport implementations that clamp their offset against the
	// current content (e.g. bubbletea's viewport) could receive syncYOffset
	// during Update while they still hold the old, smaller content. The clamped
	// value would then persist into View() even after the content grew. This test
	// uses a real viewport.Model to verify the selected picker option is visible
	// after opening the dropdown.
	selectRow := selectfield.NewFromStrings([]string{"a", "b", "c", "d", "e", "f"})
	selectRow.Set("e")

	// Content before opening: 0 header, 1 spacer, 2 select inline = 3 lines.
	// Picker adds 6 lines, total becomes 9. Committed option "e" is at picker
	// index 4, so the outer cursor line is 2 + 1 + 4 = 7.
	nav := navigator.NewBuilder().
		WithItems(testLabel("header"), testLabel("spacer"), selectRow).
		WithControllerItems(selectRow.Controller()).
		WithHeight(3).
		Build()

	vp := viewport.New()
	nav.ViewportController().SetViewport(&vp)
	_ = nav.FocusFirst()

	// Move focus onto the select field (line 2).
	sendKey(t, nav, "down")
	sendKey(t, nav, "down")
	require.Equal(t, 2, nav.FocusedIndex())

	// Open the picker and process the resulting LockFocusMsg.
	_, cmd := nav.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	require.NotNil(t, cmd)

	_, _ = nav.Update(cmd())

	require.Greater(t, nav.CursorLine(), 2)

	// Force View() to sync content and re-apply the offset.
	_ = nav.ViewportController().View()

	cursor := nav.CursorLine()
	height := nav.ViewportController().Height()

	require.Equal(t, 7, cursor)
	require.Equal(t, 3, height)

	// The controller's intended offset must match the viewport's actual offset
	// after View() has synced the new content.
	require.Equal(t, nav.ViewportController().YOffset(), vp.YOffset())
	require.GreaterOrEqual(t, cursor, vp.YOffset())
	require.Less(t, cursor, vp.YOffset()+height)
}
