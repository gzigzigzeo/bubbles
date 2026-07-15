package toggle

import (
	tea "charm.land/bubbletea/v2"

	"github.com/gzigzigzeo/bubbles/navigator/row"
)

// DisableController disables or enables a set of dependent rows when it sees
// toggle state messages from its configured source toggle.
type DisableController struct {
	source  *Model
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

// WithSource binds the controller to a specific toggle. It only reacts to
// [OnMsg] and [OffMsg] messages whose [Source] pointer matches the given
// toggle.
func WithSource(source *Model) DisableControllerOption {
	return func(c *DisableController) {
		c.source = source
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

// NewDisableControllerFor creates a controller bound to the given source toggle.
func NewDisableControllerFor(source *Model, targets ...row.Disableable) *DisableController {
	return NewDisableController(targets, WithSource(source))
}

// Update reacts to toggle state messages and returns commands that disable or
// enable the controlled targets. It ignores all other messages.
func (c *DisableController) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case OnMsg:
		if c.matchesSource(msg.Source) {
			return c.sync(true)
		}
	case OffMsg:
		if c.matchesSource(msg.Source) {
			return c.sync(false)
		}
	}

	return nil
}

// matchesSource reports whether the message source matches the configured
// source. A nil source accepts any message for backward compatibility.
func (c *DisableController) matchesSource(source *Model) bool {
	if c.source == nil {
		return true
	}

	return c.source == source
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

// Items returns the rows controlled by this controller as tea models. It is
// empty because the controlled targets live in the outer navigator and are
// updated through messages.
func (c *DisableController) Items() []tea.Model {
	return nil
}
