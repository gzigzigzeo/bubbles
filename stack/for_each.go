package stack

import tea "charm.land/bubbletea/v2"

// UpdateFunc transforms an item into its updated model and an optional command.
type UpdateFunc[M tea.Model] func(M) (M, tea.Cmd)

// ForEach visits every non-nil item, replaces it with the model returned by fn,
// and returns the updated slice with the collected commands as a [tea.Sequence].
func ForEach[M tea.Model](items []M, fn UpdateFunc[M]) ([]M, tea.Cmd) {
	cmds := make([]tea.Cmd, 0, len(items))

	for i, it := range items {
		if any(it) == nil {
			continue
		}

		updated, cmd := fn(it)
		items[i] = updated
		cmds = append(cmds, cmd)
	}

	return items, tea.Sequence(cmds...)
}
