package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestSearch_ForwardBasic(t *testing.T) {
	ed := newTestEditor(t, "hello world hello")

	// Enter search mode with /
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone))

	if ed.mode != ModeSearch {
		t.Fatalf("expected ModeSearch, got %v", ed.mode)
	}

	// Type "world"
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'w', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))

	// Press Enter to search
	ed.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))

	if ed.mode != ModeNormal {
		t.Fatalf("expected ModeNormal after search, got %v", ed.mode)
	}

	// Cursor should be at 'w' in "world" (position 6)
	if ed.cx != 6 {
		t.Fatalf("expected cursor at position 6, got %d", ed.cx)
	}
}

func TestSearch_BackwardBasic(t *testing.T) {
	ed := newTestEditor(t, "foo bar foo baz")

	// Move cursor to end
	ed.cx = 14

	// Enter backward search with ?
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '?', tcell.ModNone))

	if !ed.searchForward {
		// searchForward should be false for backward search
	}

	// Type "foo"
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))

	// Press Enter
	ed.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))

	// Should find "foo" at position 8 (second occurrence, searching backward)
	if ed.cx != 8 {
		t.Fatalf("expected cursor at position 8, got %d", ed.cx)
	}
}

func TestSearch_RepeatForward(t *testing.T) {
	ed := newTestEditor(t, "cat dog cat mouse cat")

	// Search for "cat"
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))

	// First match at position 0
	if ed.cx != 0 {
		t.Fatalf("first match: expected cursor at 0, got %d", ed.cx)
	}

	// Press 'n' to find next
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))

	// Second match at position 8
	if ed.cx != 8 {
		t.Fatalf("second match: expected cursor at 8, got %d", ed.cx)
	}

	// Press 'n' again
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))

	// Third match at position 18
	if ed.cx != 18 {
		t.Fatalf("third match: expected cursor at 18, got %d", ed.cx)
	}
}

func TestSearch_RepeatBackward(t *testing.T) {
	ed := newTestEditor(t, "one two one three one")

	// Move to end
	ed.cx = 20

	// Search backward for "one"
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '?', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))

	// Should find last "one" at position 18
	if ed.cx != 18 {
		t.Fatalf("first match: expected cursor at 18, got %d", ed.cx)
	}

	// Press 'n' to continue backward
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))

	// Should find middle "one" at position 8
	if ed.cx != 8 {
		t.Fatalf("second match: expected cursor at 8, got %d", ed.cx)
	}
}

func TestSearch_ReverseDirection(t *testing.T) {
	ed := newTestEditor(t, "alpha beta alpha gamma")

	// Forward search for "alpha"
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))

	// At first "alpha" (position 0)
	if ed.cx != 0 {
		t.Fatalf("expected cursor at 0, got %d", ed.cx)
	}

	// Press 'n' to go to next
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))

	// At second "alpha" (position 11)
	if ed.cx != 11 {
		t.Fatalf("expected cursor at 11, got %d", ed.cx)
	}

	// Press 'N' to go back (reverse direction)
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'N', tcell.ModNone))

	// Should be back at first "alpha" (position 0)
	if ed.cx != 0 {
		t.Fatalf("after N: expected cursor at 0, got %d", ed.cx)
	}
}

func TestSearch_NoMatch(t *testing.T) {
	ed := newTestEditor(t, "hello world")

	// Search for something that doesn't exist
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))

	// Cursor should not move (stay at 0)
	if ed.cx != 0 {
		t.Fatalf("cursor should not move on no match, got %d", ed.cx)
	}

	// Status should indicate not found
	if ed.statusMsg != "pattern not found: xyz" {
		t.Fatalf("expected 'pattern not found' message, got %q", ed.statusMsg)
	}
}

func TestSearch_Wraparound(t *testing.T) {
	ed := newTestEditor(t, "start middle end")

	// Move to end
	ed.cx = 15

	// Search forward (should wrap to beginning)
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))

	// Should wrap to position 0
	if ed.cx != 0 {
		t.Fatalf("expected wraparound to position 0, got %d", ed.cx)
	}

	if ed.statusMsg != "search wrapped" {
		t.Fatalf("expected 'search wrapped' message, got %q", ed.statusMsg)
	}
}

func TestSearch_EscapeCancel(t *testing.T) {
	ed := newTestEditor(t, "test content")

	// Enter search mode
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone))

	if ed.mode != ModeSearch {
		t.Fatalf("expected ModeSearch, got %v", ed.mode)
	}

	// Type something
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))

	// Press Esc to cancel
	ed.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	if ed.mode != ModeNormal {
		t.Fatalf("expected ModeNormal after Esc, got %v", ed.mode)
	}

	// Cursor should not have moved
	if ed.cx != 0 {
		t.Fatalf("cursor should not move on cancel, got %d", ed.cx)
	}
}

func TestSearch_Backspace(t *testing.T) {
	ed := newTestEditor(t, "testing")

	// Enter search mode
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone))

	// Type "abcd"
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))

	// Backspace twice
	ed.handleKey(tcell.NewEventKey(tcell.KeyBackspace, 0, tcell.ModNone))
	ed.handleKey(tcell.NewEventKey(tcell.KeyBackspace, 0, tcell.ModNone))

	// Buffer should be "ab"
	if string(ed.searchBuf) != "ab" {
		t.Fatalf("expected searchBuf 'ab', got %q", string(ed.searchBuf))
	}
}
