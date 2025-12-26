package editor

import (
	"strings"
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
func TestCompletionSorting_Stable(t *testing.T) {
	// Create buffer with words of varying lengths starting with "te"
	e := newTestEditor(t, "test testing tested tester terminology tea")

	// Type "te" at end
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))

	// Trigger completion
	e.handleKey(tcell.NewEventKey(tcell.KeyCtrlN, 0, tcell.ModNone))

	if !e.completionActive {
		t.Fatal("completion should be active")
	}

	// Verify stable sorting: length first (shorter first), then alphabetical
	// Expected order: tea, test, tester, tested, testing, terminology
	expected := []string{"tea", "test", "tested", "tester", "testing", "terminology"}

	if len(e.completionCandidates) != len(expected) {
		t.Errorf("expected %d candidates, got %d: %v", len(expected), len(e.completionCandidates), e.completionCandidates)
	}

	for i, word := range expected {
		if i >= len(e.completionCandidates) {
			break
		}
		if e.completionCandidates[i] != word {
			t.Errorf("candidate %d: expected %q, got %q\nAll: %v", i, word, e.completionCandidates[i], e.completionCandidates)
		}
	}

	// Run twice to verify deterministic ordering
	e2 := newTestEditor(t, "test testing tested tester terminology tea")
	e2.handleKey(tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone))
	e2.handleKey(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	e2.handleKey(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	e2.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	e2.handleKey(tcell.NewEventKey(tcell.KeyCtrlN, 0, tcell.ModNone))

	if len(e.completionCandidates) != len(e2.completionCandidates) {
		t.Error("completion results should be deterministic")
	}

	for i := range e.completionCandidates {
		if e.completionCandidates[i] != e2.completionCandidates[i] {
			t.Errorf("inconsistent ordering at index %d: %q vs %q",
				i, e.completionCandidates[i], e2.completionCandidates[i])
		}
	}
}

func TestCompletionPopup_Highlighting(t *testing.T) {
	e := newTestEditor(t, "hello help helicopter")

	// Type "hel" at end
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'l', tcell.ModNone))

	// Trigger completion
	e.handleKey(tcell.NewEventKey(tcell.KeyCtrlN, 0, tcell.ModNone))

	if !e.completionActive {
		t.Fatal("completion should be active")
	}

	if !e.popupActive {
		t.Fatal("popup should be active")
	}

	// Check that popup lines contain highlighting brackets
	foundHighlight := false
	for _, line := range e.popupLines {
		if strings.Contains(line, "[hel]") {
			foundHighlight = true
			break
		}
	}

	if !foundHighlight {
		t.Errorf("expected popup lines to contain [hel] highlighting, got: %v", e.popupLines)
	}

	// First candidate should be selected (with ">")
	if len(e.popupLines) > 0 && !strings.HasPrefix(e.popupLines[0], "> ") {
		t.Errorf("first line should start with '> ', got: %q", e.popupLines[0])
	}
}
func TestUndoGrouping_MultipleLines(t *testing.T) {
	e := newTestEditor(t, "")

	// Enter insert mode and type multiple lines
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'c', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'r', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	text := e.buffer.String()
	if text != "first\nsecond\nthird" {
		t.Fatalf("expected 'first\\nsecond\\nthird', got %q", text)
	}

	// Undo should remove "third"
	e.undo()
	text = e.buffer.String()
	if text != "first\nsecond" {
		t.Errorf("expected 'first\\nsecond' after first undo, got %q", text)
	}

	// Undo should remove "second" (newline was already removed with "third")
	e.undo()
	text = e.buffer.String()
	if text != "first" {
		t.Errorf("expected 'first' after second undo, got %q", text)
	}

	// Undo should remove "first" and newline
	e.undo()
	text = e.buffer.String()
	if text != "" {
		t.Errorf("expected empty after third undo, got %q", text)
	}
}

func TestUndoGrouping_WithBackspace(t *testing.T) {
	e := newTestEditor(t, "")

	// Enter insert mode and type with backspaces
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyBackspace2, 0, tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	text := e.buffer.String()
	if text != "test" {
		t.Fatalf("expected 'test', got %q", text)
	}

	// Undo should remove entire editing session including backspace
	e.undo()
	text = e.buffer.String()
	if text != "" {
		t.Errorf("expected empty after undo, got %q", text)
	}
}

func TestUndoGrouping_CompletionInteraction(t *testing.T) {
	e := newTestEditor(t, "hello world")

	// Enter insert mode at end
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone))

	// Trigger completion
	e.handleKey(tcell.NewEventKey(tcell.KeyCtrlN, 0, tcell.ModNone))
	if !e.completionActive {
		t.Fatal("completion should be active")
	}

	// Exit insert mode (completion should be part of undo group)
	e.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	text := e.buffer.String()
	if !strings.HasPrefix(text, "hello world hel") {
		t.Fatalf("expected completion applied, got %q", text)
	}

	// Undo should remove entire insert session including completion
	e.undo()
	text = e.buffer.String()
	if text != "hello world" {
		t.Errorf("expected 'hello world' after undo, got %q", text)
	}
}

func TestUndoGrouping_MultipleEnters(t *testing.T) {
	e := newTestEditor(t, "")

	// Enter insert mode and type text with multiple newlines
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))

	text := e.buffer.String()
	if text != "a\n\nb" {
		t.Fatalf("expected 'a\\n\\nb', got %q", text)
	}

	// First undo removes "b"
	e.undo()
	text = e.buffer.String()
	if text != "a\n" {
		t.Errorf("expected 'a\\n' after first undo, got %q", text)
	}

	// Second undo removes empty line (second newline)
	e.undo()
	text = e.buffer.String()
	if text != "a" {
		t.Errorf("expected 'a' after second undo, got %q", text)
	}

	// Third undo removes "a" and first newline
	e.undo()
	text = e.buffer.String()
	if text != "" {
		t.Errorf("expected empty after third undo, got %q", text)
	}
}
