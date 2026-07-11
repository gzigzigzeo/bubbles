package menu_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/menu"
)

var (
	keyDown  = tea.KeyPressMsg{Text: "j", Code: 'j'}
	keyUp    = tea.KeyPressMsg{Text: "k", Code: 'k'}
	keyEnter = tea.KeyPressMsg{Code: tea.KeyEnter}
)

func options(names ...string) []menu.Option[string] {
	opts := make([]menu.Option[string], len(names))
	for i, name := range names {
		opts[i] = menu.Option[string]{Name: name, Value: name}
	}

	return opts
}

func newMenu(opts []menu.Option[string]) *menu.Model[string] {
	m := menu.New(opts)
	m.SetWidth(40)

	return m
}

func pressKey(t *testing.T, m *menu.Model[string], km tea.KeyMsg) tea.Cmd {
	t.Helper()

	updated, cmd := m.Update(km)
	require.Same(t, m, updated)

	return cmd
}

func fireMsg(t *testing.T, cmd tea.Cmd) tea.Msg {
	t.Helper()
	require.NotNil(t, cmd)

	return cmd()
}

func TestCursor_upDownClamped(t *testing.T) {
	m := menu.New(options("a", "b", "c"))
	require.Equal(t, "a", m.Cursor())

	pressKey(t, m, keyUp)
	require.Equal(t, "a", m.Cursor(), "cursor must not go below the first option")

	pressKey(t, m, keyDown)
	pressKey(t, m, keyDown)
	require.Equal(t, "c", m.Cursor())

	pressKey(t, m, keyDown)
	require.Equal(t, "c", m.Cursor(), "cursor must not exceed the last option")
}

func TestSelect_firesChoiceForSelectedOption(t *testing.T) {
	m := menu.New(options("a", "b", "c"))
	pressKey(t, m, keyDown)

	cmd := pressKey(t, m, keyEnter)
	msg := fireMsg(t, cmd)

	choice, ok := msg.(menu.ChoiceMsg[string])
	require.True(t, ok)
	require.Equal(t, "b", choice.Value)
	require.Equal(t, "b", choice.Option.Name)
}

func TestSetCursor_movesToMatchingValueAndScrollsIntoView(t *testing.T) {
	m := newMenu(options("a", "b", "c", "d", "e"))
	m.SetHeight(2)

	m.SetValue("e")
	require.Equal(t, "e", m.Cursor())
	require.Contains(t, m.View().Content, "e", "cursor's row must be scrolled into view")
	require.NotContains(t, m.View().Content, "a", "row scrolled out of the 2-row window must not still render")
}

func TestSetCursor_noopWhenValueNotFound(t *testing.T) {
	m := newMenu(options("a", "b", "c"))

	m.SetValue("b")
	require.Equal(t, "b", m.Cursor())

	m.SetValue("does-not-exist")
	require.Equal(t, "b", m.Cursor(), "cursor must stay put when no option matches the value")
}

func TestMarker_showsDistinctFromCursor(t *testing.T) {
	m := newMenu(options("a", "b", "c"))
	m.SetStyles(menu.Styles{CursorMarked: lipgloss.NewStyle().SetString("m")})
	m.SetMarker("b")

	view := m.View().Content
	require.Contains(t, view, "m", "marked row should show the marker while the cursor is elsewhere")
}

func TestMarker_cursorTakesPrecedence(t *testing.T) {
	m := newMenu(options("a", "b", "c"))
	m.SetStyles(menu.Styles{
		CursorFocused: lipgloss.NewStyle().SetString("f"),
		CursorMarked:  lipgloss.NewStyle().SetString("m"),
	})
	m.SetMarker("a")

	require.Contains(t, m.View().Content, "f", "cursor glyph should still show on the marked row when the cursor is on it")
	require.NotContains(t, m.View().Content, "m")
}

func TestMarker_absentUntilSet(t *testing.T) {
	m := newMenu(options("a", "b", "c"))
	m.SetStyles(menu.Styles{CursorMarked: lipgloss.NewStyle().SetString("m")})

	require.NotContains(t, m.View().Content, "m")
}

func arrowStyles() menu.Styles {
	return menu.Styles{
		ScrollUp:   lipgloss.NewStyle().SetString("^"),
		ScrollDown: lipgloss.NewStyle().SetString("v"),
	}
}

func TestView_noScrollArrowsWhenEverythingFits(t *testing.T) {
	m := menu.New(options("a", "b", "c"))
	m.SetStyles(arrowStyles())
	m.SetWidth(40)

	view := m.View()
	require.Contains(t, view.Content, "a")
	require.NotContains(t, view.Content, "^")
	require.NotContains(t, view.Content, "v")
}

func TestView_showsScrollDownArrowWhenMoreOptionsBelow(t *testing.T) {
	m := menu.New(options("a", "b", "c", "d", "e"))
	m.SetStyles(arrowStyles())
	m.SetWidth(40)
	m.SetHeight(2)

	view := m.View()
	require.Contains(t, view.Content, "v", "down arrow expected on the last visible line when more options are below")
	require.NotContains(t, view.Content, "^", "nothing scrolled above yet, so no up arrow expected")
}

func TestView_showsScrollUpArrowOnceScrolledDown(t *testing.T) {
	m := menu.New(options("a", "b", "c", "d", "e"))
	m.SetStyles(arrowStyles())
	m.SetWidth(40)
	m.SetHeight(2)

	m.SetValue("e")

	view := m.View()
	require.Contains(t, view.Content, "^", "up arrow expected once scrolled past the first options")
}

func TestDescription_omittedWhenNoOptionHasOne(t *testing.T) {
	m := newMenu([]menu.Option[string]{
		{Name: "a", Value: "a"},
		{Name: "bb", Value: "bb"},
	})

	view := m.View()
	require.Contains(t, view.Content, "a")
	require.NotContains(t, view.Content, "  a", "no padding gutter expected without a description column")
}

func TestDescription_alignedWhenPresent(t *testing.T) {
	m := newMenu([]menu.Option[string]{
		{Name: "a", Description: "first", Value: "a"},
		{Name: "bb", Description: "second", Value: "bb"},
	})

	view := m.View()
	require.Contains(t, view.Content, "first")
	require.Contains(t, view.Content, "second")
}

func TestDescription_neverWrapped(t *testing.T) {
	desc := "a very long description that would previously have wrapped at a narrow width"
	optsFor := func() []menu.Option[string] {
		return []menu.Option[string]{{Name: "a", Description: desc, Value: "a"}}
	}

	wide := menu.New(optsFor())
	wide.SetWidth(200)
	require.Contains(t, wide.View().Content, desc, "the full description must render, uncapped, when the viewport is wide enough for it")

	narrow := menu.New(optsFor())
	narrow.SetWidth(20)
	wideLines := strings.Count(wide.View().Content, "\n")
	narrowLines := strings.Count(narrow.View().Content, "\n")
	require.Equal(t, wideLines, narrowLines, "a narrow viewport must not wrap the description onto extra lines")
}

func TestSyncContent_truncatesRowWiderThanSetWidth(t *testing.T) {
	desc := "a very long description that would previously have wrapped at a narrow width"
	m := menu.New([]menu.Option[string]{{Name: "a", Description: desc, Value: "a"}})
	m.SetWidth(20)

	view := m.View().Content
	require.Contains(t, view, "…", "a row wider than SetWidth must be truncated with a trailing ellipsis")
	require.NotContains(t, view, desc, "the full description must not survive truncation")
	require.Equal(t, 0, strings.Count(view, "\n"), "truncation must not push the row onto a second line")
}

func TestSyncContent_untruncatedWhenRowFitsWidth(t *testing.T) {
	m := newMenu([]menu.Option[string]{{Name: "a", Description: "short", Value: "a"}})

	view := m.View().Content
	require.Contains(t, view, "short")
	require.NotContains(t, view, "…", "a row that already fits SetWidth must not be truncated")
}

func TestSyncContent_untruncatedWhenWidthGenerous(t *testing.T) {
	desc := "a very long description that would previously have wrapped at a narrow width"
	m := menu.New([]menu.Option[string]{{Name: "a", Description: desc, Value: "a"}})
	m.SetWidth(200)

	view := m.View().Content
	require.Contains(t, view, desc, "a row narrower than SetWidth must render uncapped")
	require.NotContains(t, view, "…")
}

func TestView_totalHeightNeverExceedsConfiguredHeight(t *testing.T) {
	m := newMenu(options("a", "b", "c", "d", "e"))
	m.SetHeight(3)

	lines := strings.Count(m.View().Content, "\n") + 1
	require.LessOrEqual(t, lines, 3, "View() must not render more lines than SetHeight requested")
}

func TestCursor_scrollsByOneRowNotByFullPage(t *testing.T) {
	m := newMenu(options("a", "b", "c", "d", "e"))
	m.SetHeight(2)

	pressKey(t, m, keyDown) // -> b (still in window)
	pressKey(t, m, keyDown) // -> c (one past the bottom edge)

	content := m.View().Content
	require.Contains(t, content, "b", "the previously visible row must stay visible after a one-row scroll")
	require.Contains(t, content, "c")
	require.NotContains(t, content, "a")
	require.NotContains(t, content, "d", "must not jump a full page ahead when moving one row past the bottom edge")
}
