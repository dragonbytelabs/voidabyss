package editor

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
)

func isCtrlR(k *tcell.EventKey) bool {
	if k == nil {
		return false
	}

	// 1) Some terminals/tcell normalize Ctrl+R into the "control keycode" 0x12 (18).
	// This is independent of rune/mod.
	if k.Key() == tcell.Key(0x12) {
		return true
	}

	// 2) tcell native constant (keep it, but don't rely on it exclusively)
	if k.Key() == tcell.KeyCtrlR {
		return true
	}

	// 3) Some inputs send Ctrl+r as rune 'r' with ModCtrl.
	if k.Key() == tcell.KeyRune && (k.Modifiers()&tcell.ModCtrl) != 0 {
		if k.Rune() == 'r' || k.Rune() == 'R' {
			return true
		}
	}

	// 4) Some terminals encode Ctrl+R as a rune 0x12.
	if k.Key() == tcell.KeyRune && k.Rune() == rune(0x12) {
		return true
	}

	return false
}

func isCtrlN(k *tcell.EventKey) bool {
	if k == nil {
		return false
	}
	// Ctrl+N is ASCII 0x0E (14)
	if k.Key() == tcell.Key(0x0E) || k.Key() == tcell.KeyCtrlN {
		return true
	}
	if k.Key() == tcell.KeyRune && (k.Modifiers()&tcell.ModCtrl) != 0 {
		if k.Rune() == 'n' || k.Rune() == 'N' {
			return true
		}
	}
	if k.Key() == tcell.KeyRune && k.Rune() == rune(0x0E) {
		return true
	}
	return false
}

func isCtrlP(k *tcell.EventKey) bool {
	if k == nil {
		return false
	}
	// Ctrl+P is ASCII 0x10 (16)
	if k.Key() == tcell.Key(0x10) || k.Key() == tcell.KeyCtrlP {
		return true
	}
	if k.Key() == tcell.KeyRune && (k.Modifiers()&tcell.ModCtrl) != 0 {
		if k.Rune() == 'p' || k.Rune() == 'P' {
			return true
		}
	}
	if k.Key() == tcell.KeyRune && k.Rune() == rune(0x10) {
		return true
	}
	return false
}

func isCtrlO(k *tcell.EventKey) bool {
	if k == nil {
		return false
	}
	// Ctrl+O is ASCII 0x0F (15)
	if k.Key() == tcell.Key(0x0F) || k.Key() == tcell.KeyCtrlO {
		return true
	}
	if k.Key() == tcell.KeyRune && (k.Modifiers()&tcell.ModCtrl) != 0 {
		if k.Rune() == 'o' || k.Rune() == 'O' {
			return true
		}
	}
	if k.Key() == tcell.KeyRune && k.Rune() == rune(0x0F) {
		return true
	}
	return false
}

func isCtrlI(k *tcell.EventKey) bool {
	if k == nil {
		return false
	}
	// Ctrl+I is ASCII 0x09 (9) - same as Tab
	// We need to be careful here as Tab is commonly used
	if k.Key() == tcell.Key(0x09) || k.Key() == tcell.KeyCtrlI {
		return true
	}
	if k.Key() == tcell.KeyRune && (k.Modifiers()&tcell.ModCtrl) != 0 {
		if k.Rune() == 'i' || k.Rune() == 'I' {
			return true
		}
	}
	if k.Key() == tcell.KeyRune && k.Rune() == rune(0x09) {
		return true
	}
	return false
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func isSpace(r rune) bool { return unicode.IsSpace(r) }

// "word" for lowercase motions (letters/digits/_)
func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_'
}

// "WORD" for uppercase motions: any non-space run
func isWORDChar(r rune) bool {
	return !isSpace(r)
}

func isWordCharSmall(r rune) bool {
	// your existing definition (letters/digits/_)
	return isWordChar(r)
}

func isWordCharBig(r rune) bool {
	// Vim "WORD": any run of non-whitespace
	return !isSpace(r)
}
