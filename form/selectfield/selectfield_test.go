package selectfield_test

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/stretchr/testify/require"

	"github.com/gzigzigzeo/bubbles/form/selectfield"
	"github.com/gzigzigzeo/bubbles/menu"
)

// pickerStyles returns a distinctive, fixed-width style set: ◀/▶/▲/▼ all
// render at the same 2-column width, matching the real app's convention
// (see ui/styles.go) and letting tests locate each glyph unambiguously. The
// same style set is used for all three (Focused/Blurred/Disabled) variants
// since these tests only ever exercise the focused state.
func pickerStyles() selectfield.Styles {
	s := selectfield.SelectStyles{
		Value:      lipgloss.NewStyle(),
		ArrowLeft:  lipgloss.NewStyle().SetString("◀ "),
		ArrowRight: lipgloss.NewStyle().SetString(" ▶"),
		Picker: menu.Styles{
			ScrollUp:      lipgloss.NewStyle().SetString("▲ "),
			ScrollDown:    lipgloss.NewStyle().SetString("▼ "),
			CursorFocused: lipgloss.NewStyle().SetString("▶ "),
			CursorBlurred: lipgloss.NewStyle().SetString("  "),
			CursorMarked:  lipgloss.NewStyle().SetString("✓ "),
			LabelFocused:  lipgloss.NewStyle(),
			LabelBlurred:  lipgloss.NewStyle(),
		},
	}

	return selectfield.Styles{Focused: s, Blurred: s, Disabled: s}
}

var (
	keyDown  = tea.KeyPressMsg{Text: "j", Code: 'j'}
	keyEnter = tea.KeyPressMsg{Code: tea.KeyEnter}
	keyEsc   = tea.KeyPressMsg{Code: tea.KeyEscape}
)

func newFocusedField() *selectfield.Model[string] {
	f := selectfield.NewFromStrings([]string{"a", "b", "c"})
	f.Focus()

	return f
}

func TestSelectField_getSetRoundtrip(t *testing.T) {
	f := selectfield.NewFromStrings([]string{"a", "b", "c"})
	require.Equal(t, "a", f.Get(), "first option should be committed by default")

	f.Set("b")
	require.Equal(t, "b", f.Get())

	f.Set("does-not-exist")
	require.Equal(t, "b", f.Get(), "Set with an unknown value must be a no-op")
}

func TestSelectField_setOptionsKeepsCommittedValueWhenStillPresent(t *testing.T) {
	f := selectfield.NewFromStrings([]string{"a", "b", "c"})
	f.Set("b")

	f.SetOptions([]selectfield.Option[string]{
		{Value: "b", Label: "b"},
		{Value: "d", Label: "d"},
	})

	require.Equal(t, "b", f.Get(), "committed value present in the new options must be kept")
}

func TestSelectField_setOptionsResetsToFirstWhenCommittedValueGone(t *testing.T) {
	f := selectfield.NewFromStrings([]string{"a", "b", "c"})
	f.Set("b")

	f.SetOptions([]selectfield.Option[string]{
		{Value: "x", Label: "x"},
		{Value: "y", Label: "y"},
	})

	require.Equal(t, "x", f.Get(), "committed value absent from the new options must reset to the first one")
}

func TestSelectField_openNavigateSelectCommits(t *testing.T) {
	f := newFocusedField()

	_, _ = f.Update(keyEnter) // open the picker
	_, _ = f.Update(keyDown)  // move to "b"
	_, _ = f.Update(keyEnter) // commit

	require.Equal(t, "b", f.Get())
}

func TestSelectField_escapeCancelsWithoutCommitting(t *testing.T) {
	f := newFocusedField()

	_, _ = f.Update(keyEnter) // open the picker
	_, _ = f.Update(keyDown)  // move to "b"
	_, _ = f.Update(keyEsc)   // cancel

	require.Equal(t, "a", f.Get(), "Escape must not commit the highlighted option")
}

// TestSelectField_pickerRespectsAvailableHeightCeiling guards against
// pickerVisible() undercounting its own inline row and the blank lines it
// reserves around the dropdown (see pickerMarginTop/pickerMarginBottom):
// when the option count is capped to fit availableHeight, the field's
// rendered View() must never exceed that ceiling, otherwise a dropdown's
// bottom rows silently overflow past whatever is clipping it.
func TestSelectField_pickerRespectsAvailableHeightCeiling(t *testing.T) {
	opts := make([]string, 20)
	for i := range opts {
		opts[i] = fmt.Sprintf("opt%d", i)
	}

	f := selectfield.NewFromStrings(opts)
	f.Focus()

	const ceiling = 5
	f.SetAvailableHeight(ceiling)

	_, _ = f.Update(keyEnter) // open the picker

	require.LessOrEqual(t, lipgloss.Height(f.View().Content), ceiling,
		"rendered picker must never exceed the available height ceiling")
}

// TestSelectField_pickerCursorAlignsWithClosedArrow guards the actual bug
// this field's LeftGutter/pickerRows split fixes: menu.Model bundles its
// scroll indicator and cursor glyph into one composite gutter column per
// row, which used to leave the picker's ▶ two columns right of where the
// closed field's own ◀ sits. Peeling the scroll column off into LeftGutter
// should leave ▶ flush at the same column as ◀.
func TestSelectField_pickerCursorAlignsWithClosedArrow(t *testing.T) {
	f := selectfield.NewFromStrings([]string{"a", "b", "c"})
	f.Focus()
	f.SetStyles(pickerStyles())
	f.SetWidth(40)

	_, _ = f.Update(keyEnter) // open the picker; cursor starts on "a"

	lines := strings.Split(f.View().Content, "\n")
	arrowCol := strings.Index(lines[0], "◀")
	require.GreaterOrEqual(t, arrowCol, 0, "closed inline view must show the left arrow")

	cursorCol := -1
	for _, l := range lines[1:] {
		if idx := strings.Index(l, "▶"); idx >= 0 {
			cursorCol = idx
			break
		}
	}
	require.GreaterOrEqual(t, cursorCol, 0, "an open picker must show the cursor glyph on some row")
	require.Equal(t, arrowCol, cursorCol, "picker cursor glyph must land in the same column as the closed field's arrow")
}

// TestSelectField_leftGutterLineCountMatchesView guards the invariant
// LeftGutterAware documents: LeftGutter() must render exactly one line per
// line of View().Content, in both the closed and open states, or form.Model's
// lipgloss.JoinHorizontal silently desyncs the two columns.
func TestSelectField_leftGutterLineCountMatchesView(t *testing.T) {
	f := selectfield.NewFromStrings([]string{"a", "b", "c", "d", "e"})
	f.Focus()
	f.SetStyles(pickerStyles())
	f.SetWidth(40)

	require.Equal(t, lipgloss.Height(f.View().Content), lipgloss.Height(f.LeftGutter()),
		"line counts must match while closed")

	_, _ = f.Update(keyEnter) // open the picker

	require.Equal(t, lipgloss.Height(f.View().Content), lipgloss.Height(f.LeftGutter()),
		"line counts must match while open")
}

// TestSelectField_leftGutterShowsScrollArrows guards that the scroll
// indicator actually reaches LeftGutter() rather than being dropped
// entirely once peeled off the picker's own rows.
func TestSelectField_leftGutterShowsScrollArrows(t *testing.T) {
	opts := make([]string, 5)
	for i := range opts {
		opts[i] = fmt.Sprintf("opt%d", i)
	}

	f := selectfield.NewFromStrings(opts)
	f.Focus()
	f.SetStyles(pickerStyles())
	f.SetWidth(40)
	f.SetAvailableHeight(4) // forces a 2-row picker window, i.e. scrolling

	_, _ = f.Update(keyEnter) // open; cursor on "opt0"

	require.NotContains(t, f.View().Content, "▼", "scroll arrow must not remain in the field's own column")
	require.Contains(t, f.LeftGutter(), "▼", "scroll-down arrow must appear in the peeled-off gutter instead")
}

// TestSelectField_setWidthCompensatesForGutter guards the load-bearing fix
// in SetWidth: form.Model has already carved LeftGutter's width out of the
// value it passes down, so SetWidth must add scrollWidth() back before
// forwarding to the menu, or the picker's rows render narrower than the
// width form.Model actually allocated to this field.
func TestSelectField_setWidthCompensatesForGutter(t *testing.T) {
	f := selectfield.NewFromStrings([]string{strings.Repeat("x", 60)})
	f.Focus()
	f.SetStyles(pickerStyles())

	const w = 20
	f.SetWidth(w)

	_, _ = f.Update(keyEnter) // open the picker

	lines := strings.Split(f.View().Content, "\n")

	var dropdownLine string
	for _, l := range lines[1:] { // skip the inline row - its ArrowRight glyph is also "▶"
		if strings.Contains(l, "▶") {
			dropdownLine = l
			break
		}
	}
	require.NotEmpty(t, dropdownLine, "an open picker must render its (only) row, carrying the cursor glyph")
	require.Equal(t, w, lipgloss.Width(dropdownLine),
		"picker row must fill the full width form.Model allocated to this field, not width-scrollWidth()")
}
