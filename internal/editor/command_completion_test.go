package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestCommandCompletion(t *testing.T) {
	e := newTestEditor(t, "test")

	// Enter command mode and type partial command
	e.mode = ModeCommand
	e.cmdBuf = []rune("fold")

	// Press Tab - should complete to first match
	e.handleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))

	// Should get "fold" (exact match comes first in sorted order)
	if got := string(e.cmdBuf); got != "fold" {
		t.Errorf("First completion: got %q, want 'fold'", got)
	}

	// Press Tab again - should get next match
	e.handleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))

	if got := string(e.cmdBuf); got != "foldall" {
		t.Errorf("Second completion: got %q, want 'foldall'", got)
	}

	// Continue cycling through completions
	e.handleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	if got := string(e.cmdBuf); got != "foldclose" {
		t.Errorf("Third completion: got %q, want 'foldclose'", got)
	}
}

func TestCommandCompletionCycle(t *testing.T) {
	e := newTestEditor(t, "test")

	e.mode = ModeCommand
	e.cmdBuf = []rune("fo")

	// Count how many completions we get
	completions := []string{}

	// First tab
	e.handleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	completions = append(completions, string(e.cmdBuf))

	// Keep pressing tab until we cycle back to original
	for i := 0; i < 10; i++ {
		e.handleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
		current := string(e.cmdBuf)
		if current == "fo" {
			// Cycled back to original
			break
		}
		completions = append(completions, current)
	}

	// Should have cycled through "fo", "fold", "foldall", "foldclose", "foldinfo", "foldopen"
	if len(completions) < 2 {
		t.Errorf("Expected multiple completions, got %d: %v", len(completions), completions)
	}

	// Should have returned to original
	if got := string(e.cmdBuf); got != "fo" {
		t.Errorf("After cycling, cmdBuf = %q, want 'fo'", got)
	}
}

func TestCommandCompletionBackward(t *testing.T) {
	e := newTestEditor(t, "test")

	e.mode = ModeCommand
	e.cmdBuf = []rune("fold")

	// Press Tab to start completion
	e.handleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	first := string(e.cmdBuf)

	// Press Shift+Tab (BackTab) to go backward
	e.handleKey(tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone))

	// Should return to original
	if got := string(e.cmdBuf); got != "fold" {
		t.Errorf("After BackTab, cmdBuf = %q, want 'fold'", got)
	}

	// Press BackTab again - should wrap to last completion
	e.handleKey(tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone))

	// Should get last match (foldopen is last alphabetically among "fold*")
	if got := string(e.cmdBuf); got == "fold" {
		t.Errorf("After second BackTab, should not be original, got %q", got)
	}

	// Now Tab forward should give us first again
	e.handleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	if got := string(e.cmdBuf); got != first {
		t.Errorf("After Tab forward, got %q, want %q", got, first)
	}
}

func TestCommandCompletionNoMatches(t *testing.T) {
	e := newTestEditor(t, "test")

	e.mode = ModeCommand
	e.cmdBuf = []rune("xyz123")

	// Press Tab with no matches
	e.handleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))

	// Should remain unchanged
	if got := string(e.cmdBuf); got != "xyz123" {
		t.Errorf("No match completion: got %q, want 'xyz123'", got)
	}
}

func TestCommandCompletionResetOnEdit(t *testing.T) {
	e := newTestEditor(t, "test")

	e.mode = ModeCommand
	e.cmdBuf = []rune("fold")

	// Start completion
	e.handleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))

	completed := string(e.cmdBuf)

	// Type a character - should reset completion
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))

	if e.cmdCompletionIdx != -1 {
		t.Errorf("After typing, cmdCompletionIdx = %d, want -1", e.cmdCompletionIdx)
	}

	if got := string(e.cmdBuf); got != completed+"x" {
		t.Errorf("After typing, cmdBuf = %q, want %q", got, completed+"x")
	}
}

func TestCommandCompletionExactMatch(t *testing.T) {
	e := newTestEditor(t, "test")

	e.mode = ModeCommand
	e.cmdBuf = []rune("q")

	// Tab should complete even with single letter
	e.handleKey(tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone))

	// Should get "q" or "q!" as first completion
	got := string(e.cmdBuf)
	if got != "q" && got != "q!" {
		t.Errorf("Completion for 'q': got %q, want 'q' or 'q!'", got)
	}
}
