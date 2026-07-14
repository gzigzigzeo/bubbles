package form_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/form"
	"github.com/gzigzigzeo/bubbles/form/button"
	"github.com/gzigzigzeo/bubbles/form/buttonstack"
	"github.com/gzigzigzeo/bubbles/form/field"
	"github.com/gzigzigzeo/bubbles/form/selectfield"
	"github.com/gzigzigzeo/bubbles/form/togglefield"
	"github.com/gzigzigzeo/bubbles/menu"
)

var (
	keyUp    = tea.KeyPressMsg{Code: tea.KeyUp}
	keyDown  = tea.KeyPressMsg{Code: tea.KeyDown}
	keyEnter = tea.KeyPressMsg{Code: tea.KeyEnter}
)

// TestForm_openPickerScrollsFullExtentIntoView guards against syncScroll
// narrowing scroll to only the cursor line when the full extent still fits.
func TestForm_openPickerScrollsFullExtentIntoView(t *testing.T) {
	opts := make([]form.Option, 0, 6)
	for range 5 {
		opts = append(opts, form.WithEntry(form.NewField("spacer", togglefield.New())))
	}

	sel := selectfield.NewFromStrings([]string{"a", "b", "c"})
	opts = append(opts, form.WithEntry(form.NewField("select", sel)))

	f := form.New(opts...)
	f.SetWidth(40)
	f.SetHeight(8)

	for range 5 {
		f.Update(keyDown) // move focus down to the selectfield.Model
	}

	f.Update(keyEnter) // open the picker; cursor starts on "a" (index 0)

	require.Contains(t, f.View(), "c",
		"opening the picker must scroll its full extent into view, not just the cursor line")
}

// TestForm_selectFieldScrollArrowRendersInGutterColumn verifies the scroll
// indicator renders in the gutter column, not merged into the field content.
func TestForm_selectFieldScrollArrowRendersInGutterColumn(t *testing.T) {
	opts := make([]string, 6)
	for i := range opts {
		opts[i] = string(rune('a' + i))
	}

	sel := selectfield.NewFromStrings(opts)

	f := form.New(form.WithEntry(form.NewField("select", sel)))
	f.SetStyles(form.Styles{
		EmptyGutter:   lipgloss.NewStyle().Width(2),
		CursorFocused: lipgloss.NewStyle().SetString("> "),
		CursorBlurred: lipgloss.NewStyle().SetString("  "),
	})
	selStyles := selectfield.SelectStyles{
		ArrowLeft:  lipgloss.NewStyle().SetString("◀ "),
		ArrowRight: lipgloss.NewStyle().SetString(" ▶"),
		Picker: menu.Styles{
			ScrollUp:      lipgloss.NewStyle().SetString("▲ "),
			ScrollDown:    lipgloss.NewStyle().SetString("▼ "),
			CursorFocused: lipgloss.NewStyle().SetString("▶ "),
			CursorBlurred: lipgloss.NewStyle().SetString("  "),
		},
	}
	sel.SetStyles(selectfield.Styles{Focused: selStyles, Blurred: selStyles, Disabled: selStyles})
	f.SetWidth(40)
	f.SetHeight(6) // small enough that pickerVisible() caps below 6 options, forcing a scroll arrow
	f.Init()       // focuses the (only) entry, so its inline arrows render

	f.Update(keyEnter) // open the picker

	lines := strings.Split(f.View(), "\n")
	arrowCol := strings.Index(lines[0], "◀")
	require.GreaterOrEqual(t, arrowCol, 0, "closed inline view must show the left arrow")

	scrollCol := -1

	for _, l := range lines[1:] {
		if idx := strings.Index(l, "▼"); idx >= 0 {
			scrollCol = idx
			break
		}
	}

	require.GreaterOrEqual(t, scrollCol, 0, "a scrolled-open picker must show a down arrow somewhere")
	require.Equal(t, arrowCol-2, scrollCol,
		"scroll arrow must render in the reserved gutter column, immediately left of the field's own edge")
}

// TestForm_optionOrderIndependent guards SetStyles/SetWidth's mutual
// recompute when WithWidth precedes WithStyles.
func TestForm_optionOrderIndependent(t *testing.T) {
	opts := make([]string, 6)
	for i := range opts {
		opts[i] = string(rune('a' + i))
	}

	sel := selectfield.NewFromStrings(opts)
	selStyles := selectfield.SelectStyles{
		ArrowLeft:  lipgloss.NewStyle().SetString("◀ "),
		ArrowRight: lipgloss.NewStyle().SetString(" ▶"),
		Picker: menu.Styles{
			ScrollUp:      lipgloss.NewStyle().SetString("▲ "),
			ScrollDown:    lipgloss.NewStyle().SetString("▼ "),
			CursorFocused: lipgloss.NewStyle().SetString("▶ "),
			CursorBlurred: lipgloss.NewStyle().SetString("  "),
		},
	}
	sel.SetStyles(selectfield.Styles{Focused: selStyles, Blurred: selStyles, Disabled: selStyles})

	f := form.New(
		form.WithEntry(form.NewField("select", sel)),
		field.WithWidth[*form.Model](40),
		form.WithStyles(form.Styles{
			EmptyGutter:   lipgloss.NewStyle().Width(2),
			CursorFocused: lipgloss.NewStyle().SetString("> "),
			CursorBlurred: lipgloss.NewStyle().SetString("  "),
		}),
	)
	f.SetHeight(6)
	f.Init()

	f.Update(keyEnter) // open the picker

	lines := strings.Split(f.View(), "\n")
	arrowCol := strings.Index(lines[0], "◀")
	require.GreaterOrEqual(t, arrowCol, 0, "closed inline view must show the left arrow")

	scrollCol := -1

	for _, l := range lines[1:] {
		if idx := strings.Index(l, "▼"); idx >= 0 {
			scrollCol = idx
			break
		}
	}

	require.GreaterOrEqual(t, scrollCol, 0, "a scrolled-open picker must show a down arrow somewhere")
	require.Equal(t, arrowCol-2, scrollCol,
		"scroll arrow must land in the gutter column even when WithWidth precedes WithStyles")
}

// TestForm_layoutOptionsBeforeEntries guards New's end-of-construction
// re-finalization when layout options are listed before WithEntry.
func TestForm_layoutOptionsBeforeEntries(t *testing.T) {
	opts := make([]string, 6)
	for i := range opts {
		opts[i] = string(rune('a' + i))
	}

	sel := selectfield.NewFromStrings(opts)
	selStyles := selectfield.SelectStyles{
		ArrowLeft:  lipgloss.NewStyle().SetString("◀ "),
		ArrowRight: lipgloss.NewStyle().SetString(" ▶"),
		Picker: menu.Styles{
			ScrollUp:      lipgloss.NewStyle().SetString("▲ "),
			ScrollDown:    lipgloss.NewStyle().SetString("▼ "),
			CursorFocused: lipgloss.NewStyle().SetString("▶ "),
			CursorBlurred: lipgloss.NewStyle().SetString("  "),
		},
	}
	sel.SetStyles(selectfield.Styles{Focused: selStyles, Blurred: selStyles, Disabled: selStyles})

	f := form.New(
		field.WithWidth[*form.Model](40),
		form.WithStyles(form.Styles{
			EmptyGutter:   lipgloss.NewStyle().Width(2),
			CursorFocused: lipgloss.NewStyle().SetString("> "),
			CursorBlurred: lipgloss.NewStyle().SetString("  "),
		}),
		form.WithEntry(form.NewField("select", sel)),
	)
	f.SetHeight(6)
	f.Init()

	f.Update(keyEnter) // open the picker

	lines := strings.Split(f.View(), "\n")
	arrowCol := strings.Index(lines[0], "◀")
	require.GreaterOrEqual(t, arrowCol, 0, "closed inline view must show the left arrow")

	scrollCol := -1

	for _, l := range lines[1:] {
		if idx := strings.Index(l, "▼"); idx >= 0 {
			scrollCol = idx
			break
		}
	}

	require.GreaterOrEqual(t, scrollCol, 0, "a scrolled-open picker must show a down arrow somewhere")
	require.Equal(t, arrowCol-2, scrollCol,
		"scroll arrow must land in the gutter column even when WithWidth/WithStyles precede WithEntry")
}

// TestForm_buttonStackVerticalNavigationExitsAtBoundary verifies that
// ↑/↓ navigate between the stack's buttons and exit to the adjacent form
// row when pressed at the first or last button.
func TestForm_buttonStackVerticalNavigationExitsAtBoundary(t *testing.T) {
	before := togglefield.New()
	bs := buttonstack.New(button.New("First", nil), button.New("Second", nil))
	after := togglefield.New()

	f := form.New(
		form.WithEntry(form.NewField("before", before)),
		form.WithEntry(bs),
		form.WithEntry(form.NewField("after", after)),
	)
	f.SetWidth(40)
	f.SetHeight(8)
	f.Init()

	f.Update(keyDown) // focus moves from "before" onto the button stack
	require.True(t, bs.Focused(), "focus should land on the button stack")
	require.Equal(t, 0, bs.Position(), "the first button should be focused initially")

	f.Update(keyDown) // move to second button within the stack
	require.True(t, bs.Focused(), "↓ within the stack must move between buttons, not leave it")
	require.Equal(t, 1, bs.Position())

	f.Update(keyUp) // move back to first button within the stack
	require.True(t, bs.Focused(), "↑ within the stack must move between buttons, not leave it")
	require.Equal(t, 0, bs.Position())

	f.Update(keyUp) // ↑ from first button exits the stack
	require.False(t, bs.Focused(), "↑ from the first button should leave the stack")

	f.Update(keyDown) // re-enter the stack from above
	require.True(t, bs.Focused())
	require.Equal(t, 0, bs.Position())

	f.Update(keyDown) // move to second button
	require.True(t, bs.Focused())
	require.Equal(t, 1, bs.Position())

	f.Update(keyDown) // ↓ from last button exits the stack
	require.False(t, bs.Focused(), "↓ from the last button should leave the stack")
}

// TestForm_pickerScrollStableAfterCursorMove guards against syncScroll
// scrolling the outer viewport on every keypress while the picker's own
// visible option window hasn't changed - see syncScroll/ensureVisible.
func TestForm_pickerScrollStableAfterCursorMove(t *testing.T) {
	options := make([]string, 20)
	for i := range options {
		options[i] = fmt.Sprintf("region%d", i)
	}

	sel := selectfield.NewFromStrings(options)

	f := form.New(
		form.WithEntry(form.NewField("marker", togglefield.New())),
		form.WithEntry(form.NewField("select", sel)),
	)
	f.SetWidth(40)
	f.SetHeight(10) // small enough that pickerVisible() clips well below 20 options

	f.Update(keyDown)  // focus the select field
	f.Update(keyEnter) // open the picker; cursor starts on region0

	before, _, _ := strings.Cut(f.View(), "\n")

	f.Update(keyDown) // move the cursor to region1, still inside the already-visible window

	after, _, _ := strings.Cut(f.View(), "\n")

	require.Equal(t, before, after,
		"moving the picker cursor within its own visible window must not scroll the outer viewport")
}
