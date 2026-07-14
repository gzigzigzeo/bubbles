package field

import (
	tea "charm.land/bubbletea/v2"
)

// FocusState manages keyboard focus over an ordered list of Controls,
// shared by form.Model and buttonstack.Model.
type FocusState struct {
	items    []Control
	position int // index into items; -1 = none
}

// NewFocusState builds a FocusState over the Control-implementing subset of
// models, in order; every other model is dropped. Call FocusFirst or
// FocusLast once construction is complete to focus an initial item.
func NewFocusState[T tea.Model](models ...T) FocusState {
	g := FocusState{position: -1}

	for _, m := range models {
		if c, ok := any(m).(Control); ok {
			g.items = append(g.items, c)
		}
	}

	return g
}

// Items returns every managed item, in order.
func (g *FocusState) Items() []Control {
	return g.items
}

// FocusFirst focuses the first non-disabled item, or none if every item is
// disabled. Call once after construction, or to preselect the entry side
// closest to where focus is entering from.
func (g *FocusState) FocusFirst() tea.Cmd {
	g.position = -1

	for i, it := range g.items {
		if !it.Disabled() {
			g.position = i
			break
		}
	}

	return nil
}

// FocusLast focuses the last non-disabled item, or none if every item is
// disabled. The mirror of FocusFirst, for focus entering from the end.
func (g *FocusState) FocusLast() tea.Cmd {
	g.position = -1

	for i := len(g.items) - 1; i >= 0; i-- {
		if !g.items[i].Disabled() {
			g.position = i
			break
		}
	}

	return nil
}

// Focus focuses the current item.
func (g *FocusState) Focus() tea.Cmd {
	if n := g.Current(); n != nil {
		return n.Focus()
	}

	return nil
}

// Blur blurs the current item.
func (g *FocusState) Blur() tea.Cmd {
	if n := g.Current(); n != nil {
		return n.Blur()
	}

	return nil
}

// Focused reports whether the current item is focused.
func (g *FocusState) Focused() bool {
	if n := g.Current(); n != nil {
		return n.Focused()
	}

	return false
}

// Current returns the currently-focused item, or nil if none is focused.
func (g *FocusState) Current() Control {
	if g.position < 0 || g.position >= len(g.items) {
		return nil
	}

	return g.items[g.position]
}

// Position returns the index of the currently-focused item, or -1 if none is focused.
func (g *FocusState) Position() int {
	return g.position
}

// shift walks focus by dir, skipping disabled items, and reports whether it
// found one. When bounded is false it wraps past either end; when true it
// stops at the boundary instead, leaving focus untouched.
func (g *FocusState) shift(dir int, bounded bool) (tea.Cmd, bool) {
	count := len(g.items)
	if count == 0 {
		return nil, false
	}

	pos := g.position

	for range count {
		pos += dir

		if bounded {
			if pos < 0 || pos >= count {
				return nil, false
			}
		} else {
			pos = ((pos % count) + count) % count
		}

		it := g.items[pos]

		if it.Disabled() {
			continue
		}

		if n := g.Current(); n != nil {
			blurCmd := n.Blur()
			g.position = pos
			enterCmd := enterFrom(it, dir)

			return tea.Batch(blurCmd, enterCmd, it.Focus()), true
		}

		g.position = pos
		enterCmd := enterFrom(it, dir)

		return tea.Batch(enterCmd, it.Focus()), true
	}

	return nil, false
}

// Shift moves focus by dir (+1 or -1), skipping disabled items and wrapping
// past either end.
func (g *FocusState) Shift(dir int) tea.Cmd {
	cmd, _ := g.shift(dir, false)

	return cmd
}

// ShiftBounded moves focus by dir like Shift, but stops at the boundary
// instead of wrapping around.
func (g *FocusState) ShiftBounded(dir int) (tea.Cmd, bool) {
	return g.shift(dir, true)
}

// Set blurs the current item (if any) and focuses item i directly,
// without skipping disabled entries.
func (g *FocusState) Set(i int) tea.Cmd {
	if i < 0 || i >= len(g.items) {
		return nil
	}

	var blurCmd tea.Cmd

	if n := g.Current(); n != nil {
		blurCmd = n.Blur()
	}

	g.position = i
	enterCmd := enterFrom(g.items[i], 1)

	return tea.Batch(blurCmd, enterCmd, g.items[i].Focus())
}

// enterFrom preselects the correct end of a FocusModel control's own nested
// focus, based on the direction focus is entering from: entering forward
// (dir >= 0) lands on the child's first item, entering backward (dir < 0)
// lands on its last. A no-op for controls that aren't FocusModel.
func enterFrom(c Control, dir int) tea.Cmd {
	n, ok := c.(FocusModel)
	if !ok {
		return nil
	}

	if dir < 0 {
		return n.FocusLast()
	}

	return n.FocusFirst()
}
