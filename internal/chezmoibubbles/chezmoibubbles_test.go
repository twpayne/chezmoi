package chezmoibubbles

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/alecthomas/assert/v2"
)

func makeKeyMsg(r rune) tea.Msg {
	switch r {
	case '\x03':
		return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	case '\r':
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case '\x1b':
		return tea.KeyPressMsg{Code: tea.KeyEsc}
	default:
		return tea.KeyPressMsg{Code: r, Text: string(r)}
	}
}

func makeKeyMsgs(s string) []tea.Msg {
	msgs := make([]tea.Msg, len(s))
	for i, r := range s {
		msgs[i] = makeKeyMsg(r)
	}
	return msgs
}

func testRunModelWithInput[M tea.Model](t *testing.T, model M, input string) M {
	t.Helper()
	for _, msg := range makeKeyMsgs(input) {
		m, _ := model.Update(msg)
		var ok bool
		model, ok = m.(M)
		assert.True(t, ok)
	}
	return model
}
