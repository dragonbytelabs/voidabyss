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

func (e *Editor) toNextWordStart(pos int, count int) int {
	if count <= 0 {
		count = 1
	}
	r := e.textRunes()
	n := len(r)
	if pos < 0 {
		pos = 0
	}
	if pos >= n {
		return pos
	}

	i := pos

	for c := 0; c < count; c++ {
		// If on a word char, consume the rest of this word
		if i < n && isWordChar(r[i]) {
			for i < n && isWordChar(r[i]) {
				i++
			}
		}

		// Consume following non-word chars (spaces/punct) up to next word start
		for i < n && !isWordChar(r[i]) {
			i++
		}
	}

	return i
}
