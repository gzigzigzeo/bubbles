package row

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type dummyStyles struct {
	name string
}

func TestStatefulStyles_StateStyles(t *testing.T) {
	styles := StateSet[dummyStyles]{
		Focused:  dummyStyles{name: "focused"},
		Blurred:  dummyStyles{name: "blurred"},
		Disabled: dummyStyles{name: "disabled"},
	}

	stateful := StatefulStyles[dummyStyles]{}
	stateful.SetStyles(styles)

	require.Equal(t, "disabled", stateful.StateStyles(true, false).name)
	require.Equal(t, "focused", stateful.StateStyles(false, true).name)
	require.Equal(t, "blurred", stateful.StateStyles(false, false).name)
	require.Equal(t, "disabled", stateful.StateStyles(true, true).name)
}
