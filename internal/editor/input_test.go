package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestIsCtrlR_KeyCtrlR(t *testing.T) {
	ev := tcell.NewEventKey(tcell.KeyCtrlR, 0, tcell.ModNone)
	if !isCtrlR(ev) {
		t.Fatalf("expected isCtrlR true for KeyCtrlR")
	}
}

func TestIsCtrlR_RuneCtrlMod(t *testing.T) {
	ev := tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModCtrl)
	if !isCtrlR(ev) {
		t.Fatalf("expected isCtrlR true for KeyRune('r') + ModCtrl")
	}
}

func TestIsCtrlR_AllEncodings(t *testing.T) {
	cases := []struct {
		name string
		ev   *tcell.EventKey
	}{
		{"KeyCtrlR", tcell.NewEventKey(tcell.KeyCtrlR, 0, tcell.ModNone)},
		{"Rune12", tcell.NewEventKey(tcell.KeyRune, rune(0x12), tcell.ModNone)},
		{"Ctrl+r", tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModCtrl)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !isCtrlR(tc.ev) {
				t.Fatalf("expected isCtrlR true; got key=%v rune=%#x mod=%v",
					tc.ev.Key(), tc.ev.Rune(), tc.ev.Modifiers())
			}
		})
	}
}
