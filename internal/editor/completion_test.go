package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestCompletion_Basic(t *testing.T) {
	e := newTestEditor(t, "hello world\nhello there\nhi world")

	// Start on "hi" line and type "hel"
	e.cy = 2
	e.cx = 0
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))

	// Press Ctrl+N to trigger completion
	e.handleKey(tcell.NewEventKey(tcell.KeyCtrlN, 0, tcell.ModNone))
	if !e.completionActive {
		t.Fatal("completion should be active")
	}

	// Should complete to "hello"
	text := e.buffer.String()
	if text[:5] != "hello" {
		t.Errorf("expected completion to 'hello', got %q", text[:10])
	}

	// Should have candidates
	if len(e.completionCandidates) == 0 {
		t.Fatal("expected completion candidates")
	}
}

func TestCompletion_Cycle(t *testing.T) {
	e := newTestEditor(t, "apple banana apricot")

	// Go to end and type " ap" to avoid merging with existing word
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone))

	// First completion (Ctrl+N)
	e.handleKey(tcell.NewEventKey(tcell.KeyCtrlN, 0, tcell.ModNone))

	if !e.completionActive {
		t.Fatal("completion should be active after first Ctrl-N")
	}

	if len(e.completionCandidates) < 2 {
		t.Fatalf("expected at least 2 candidates, got %d: %v\nBuffer: %q\nPrefix: %q",
			len(e.completionCandidates), e.completionCandidates,
			e.buffer.String(), e.completionPrefix)
	}

	candidate1 := e.completionCandidates[e.completionIndex]
	index1 := e.completionIndex

	// Cycle to next (Ctrl+N again)
	e.handleKey(tcell.NewEventKey(tcell.KeyCtrlN, 0, tcell.ModNone))

	if !e.completionActive {
		t.Fatal("completion should still be active after second Ctrl-N")
	}

	candidate2 := e.completionCandidates[e.completionIndex]
	index2 := e.completionIndex

	if candidate1 == candidate2 {
		t.Errorf("expected different candidates after cycling, both are %q\nindex1=%d index2=%d\nall candidates: %v",
			candidate1, index1, index2, e.completionCandidates)
	}
}

func TestCompletion_PreviousCycle(t *testing.T) {
	e := newTestEditor(t, "test testing tested")

	// Type "te" at end
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))

	// Start completion with Ctrl+P (previous)
	e.handleKey(tcell.NewEventKey(tcell.KeyCtrlP, 0, tcell.ModNone))
	if !e.completionActive {
		t.Fatal("completion should be active")
	}

	// Should wrap around to last candidate
	if e.completionIndex != len(e.completionCandidates)-1 {
		t.Errorf("Ctrl+P should wrap to last candidate, got index %d, candidates: %v", e.completionIndex, e.completionCandidates)
	}
}

func TestCompletion_NoMatches(t *testing.T) {
	e := newTestEditor(t, "hello world")

	// Type something with no matches
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'y', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone))

	// Try to complete
	e.handleKey(tcell.NewEventKey(tcell.KeyCtrlN, 0, tcell.ModNone))
	if e.completionActive {
		t.Errorf("completion should not activate with no matches, but got active=%v with candidates: %v", e.completionActive, e.completionCandidates)
	}

	if e.statusMsg != "no word to complete" && e.statusMsg != "no completions found" {
		t.Errorf("expected status message for no completions, got: %q", e.statusMsg)
	}
}

func TestCompletion_CancelOnEscape(t *testing.T) {
	e := newTestEditor(t, "hello world")

	// Enter insert mode and trigger completion at end
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyCtrlN, 0, tcell.ModNone))
	if !e.completionActive {
		t.Fatalf("completion should be active, candidates: %v", e.completionCandidates)
	}

	// Press Escape (exits insert mode AND cancels completion)
	e.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	if e.completionActive {
		t.Error("completion should be cancelled on Escape")
	}

	if e.popupActive {
		t.Error("popup should be closed")
	}
}

func TestCompletion_CancelOnTyping(t *testing.T) {
	e := newTestEditor(t, "hello world")

	// Enter insert mode and trigger completion at end
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyCtrlN, 0, tcell.ModNone))
	if !e.completionActive {
		t.Fatalf("completion should be active, candidates: %v", e.completionCandidates)
	}

	// Type another character
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))

	if e.completionActive {
		t.Error("completion should be cancelled when typing")
	}
}

func TestUndoGrouping_InsertSession(t *testing.T) {
	e := newTestEditor(t, "")

	// Enter insert mode
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))

	// Type multiple characters
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))

	// Exit insert mode
	e.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))
	text := e.buffer.String()
	if text != "hello" {
		t.Fatalf("expected 'hello', got %q", text)
	}

	// Undo once should remove all typed text (single undo group)
	e.undo()

	text = e.buffer.String()
	if text != "" {
		t.Errorf("expected all text to be undone in single undo, got %q", text)
	}
}

func TestUndoGrouping_NewlineBreak(t *testing.T) {
	e := newTestEditor(t, "")

	// Enter insert mode and type
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))

	// Press Enter (should break undo group)
	e.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))

	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone))

	// Exit insert mode
	e.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))
	text := e.buffer.String()
	if text != "abc\ndef" {
		t.Fatalf("expected 'abc\\ndef', got %q", text)
	}

	// First undo removes "def" and the newline
	e.undo()
	text = e.buffer.String()
	if text != "abc" {
		t.Errorf("expected 'abc' after first undo, got %q", text)
	}

	// Second undo removes "abc"
	e.undo()
	text = e.buffer.String()
	if text != "" {
		t.Errorf("expected empty after second undo, got %q", text)
	}
}

func TestBuildWordIndex(t *testing.T) {
	e := newTestEditor(t, "hello world\ntest testing\nhello again")

	// Pass exclusion range (0,0) to exclude nothing
	words := e.buildWordIndex(0, 0)

	// Should have: hello, world, test, testing, again
	expectedWords := map[string]bool{
		"hello":   true,
		"world":   true,
		"test":    true,
		"testing": true,
		"again":   true,
	}

	if len(words) != len(expectedWords) {
		t.Errorf("expected %d words, got %d: %v", len(expectedWords), len(words), words)
	}

	for _, word := range words {
		if !expectedWords[word] {
			t.Errorf("unexpected word in index: %q", word)
		}
	}
}
