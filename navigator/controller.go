package navigator

import tea "charm.land/bubbletea/v2"

// Controller is any value that participates in message handling without being
// a rendered navigator row. The navigator calls Update on each registered
// controller for every message so controllers such as toggle disable
// controllers can react to row-emitted messages.
type Controller interface {
	Update(tea.Msg) tea.Cmd
}

// itemController is the shape accepted by NavigatorBuilder.WithControllerItems:
// a controller that also exposes its rows so they can be added to the
// navigator's item list.
type itemController interface {
	Controller
	Items() []tea.Model
}
