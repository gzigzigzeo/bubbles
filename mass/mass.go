// Package mass provides helpers for forwarding messages and commands to a
// collection of child tea models.
package mass

import tea "charm.land/bubbletea/v2"

// UpdateFunc transforms an item into its updated model and an optional command.
type UpdateFunc[M tea.Model] func(M) (M, tea.Cmd)

// PropagateFunc calls a command-producing method on an item of type T.
type PropagateFunc[T tea.Model] func(T) tea.Cmd

// Update visits every non-nil item, replaces it with the model returned by the
// function, and returns the updated slice together with the collected non-nil
// commands.
func Update[M tea.Model](items []M, updateFn UpdateFunc[M]) ([]M, []tea.Cmd) {
	cmds := make([]tea.Cmd, 0, len(items))

	for idx, it := range items {
		if any(it) == nil {
			continue
		}

		updated, cmd := updateFn(it)
		items[idx] = updated

		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return items, cmds
}

// Propagate visits every non-nil item, type-asserts it to T, calls the function
// when the assertion succeeds, and returns the collected non-nil commands.
func Propagate[T tea.Model](items []tea.Model, propagateFn PropagateFunc[T]) []tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(items))

	for _, it := range items {
		if any(it) == nil {
			continue
		}

		t, ok := it.(T)
		if !ok {
			continue
		}

		if cmd := propagateFn(t); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return cmds
}
