package toggle

import (
	tea "charm.land/bubbletea/v2"

	"github.com/gzigzigzeo/bubbles/navigator/row"
)

// DisableController disables or enables a set of dependent rows when it sees
// toggle state messages.
type DisableController struct {
	targets []row.Disableable
	invert  bool
}

// DisableControllerOption configures a [DisableController].
type DisableControllerOption func(*DisableController)

// WithInvert swaps the default mapping so [OnMsg] enables targets and [OffMsg]
// disables them.
func WithInvert() DisableControllerOption {
	return func(c *DisableController) {
		c.invert = true
	}
}

// NewDisableController creates a controller for the given targets.
func NewDisableController(targets []row.Disableable, opts ...DisableControllerOption) *DisableController {
	c := &DisableController{
		targets: targets,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Update reacts to toggle state messages and returns commands that disable or
// enable the controlled targets. It ignores all other messages.
func (c *DisableController) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case OnMsg:
		return c.sync(true)
	case OffMsg:
		return c.sync(false)
	}

	return nil
}

// sync returns the commands that apply the disable/enable state matching value
// to every target.
func (c *DisableController) sync(value bool) tea.Cmd {
	if c.invert {
		value = !value
	}

	cmds := make([]tea.Cmd, 0, len(c.targets))
	for _, t := range c.targets {
		if value {
			cmds = append(cmds, t.Disable())

			continue
		}

		cmds = append(cmds, t.Enable())
	}

	return tea.Batch(cmds...)
}

// Targets returns the rows controlled by this controller.
func (c *DisableController) Targets() []row.Disableable {
	return c.targets
}
