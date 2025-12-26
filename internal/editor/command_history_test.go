package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestCommandHistory(t *testing.T) {
	e := newTestEditor(t, "test content")

	// Execute some commands
	commands := []string{"set number", "w", "q"}
	for _, cmd := range commands {
		// Enter command mode
		e.mode = ModeCommand
		e.cmdBuf = []rune(cmd)
		
		// Press Enter
		e.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	}

	// Verify history has the commands (except empty/failed ones)
	if len(e.cmdHistory) == 0 {
		t.Fatal("command history is empty")
	}

	// Enter command mode again
	e.mode = ModeCommand
	e.cmdBuf = nil

	// Press Up arrow - should get most recent command
	e.handleKey(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))
	
	lastCmd := commands[len(commands)-1]
	if got := string(e.cmdBuf); got != lastCmd {
		t.Errorf("After up arrow, cmdBuf = %q, want %q", got, lastCmd)
	}

	// Press Up again - should get second-to-last command
	e.handleKey(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))
	
	secondLastCmd := commands[len(commands)-2]
	if got := string(e.cmdBuf); got != secondLastCmd {
		t.Errorf("After second up arrow, cmdBuf = %q, want %q", got, secondLastCmd)
	}

	// Press Down - should go back to most recent
	e.handleKey(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	
	if got := string(e.cmdBuf); got != lastCmd {
		t.Errorf("After down arrow, cmdBuf = %q, want %q", got, lastCmd)
	}

	// Press Down again - should restore empty buffer
	e.handleKey(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	
	if got := string(e.cmdBuf); got != "" {
		t.Errorf("After final down arrow, cmdBuf = %q, want empty", got)
	}
}

func TestCommandHistoryDuplicates(t *testing.T) {
	e := newTestEditor(t, "test")

	// Execute same command twice
	for i := 0; i < 2; i++ {
		e.mode = ModeCommand
		e.cmdBuf = []rune("set number")
		e.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))
	}

	// Should only have one entry (duplicates are not added consecutively)
	if len(e.cmdHistory) != 1 {
		t.Errorf("cmdHistory length = %d, want 1 (no consecutive duplicates)", len(e.cmdHistory))
	}
}

func TestCommandHistoryEditing(t *testing.T) {
	e := newTestEditor(t, "test")

	// Add a command to history
	e.mode = ModeCommand
	e.cmdBuf = []rune("set number")
	e.handleKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))

	// Enter command mode and browse history
	e.mode = ModeCommand
	e.cmdBuf = nil
	e.handleKey(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))

	if string(e.cmdBuf) != "set number" {
		t.Fatalf("Expected 'set number', got %q", string(e.cmdBuf))
	}

	// Edit the command (should reset history browsing)
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone))

	if e.cmdHistoryIdx != -1 {
		t.Errorf("After editing, cmdHistoryIdx = %d, want -1", e.cmdHistoryIdx)
	}

	if got := string(e.cmdBuf); got != "set numbers" {
		t.Errorf("After editing, cmdBuf = %q, want 'set numbers'", got)
	}
}

func TestCommandHistoryEmpty(t *testing.T) {
	e := newTestEditor(t, "test")

	// Try to navigate history when empty
	e.mode = ModeCommand
	e.cmdBuf = nil

	e.handleKey(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))
	
	if len(e.cmdBuf) != 0 {
		t.Errorf("Up arrow on empty history should do nothing, got cmdBuf = %q", string(e.cmdBuf))
	}

	e.handleKey(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	
	if len(e.cmdBuf) != 0 {
		t.Errorf("Down arrow on empty history should do nothing, got cmdBuf = %q", string(e.cmdBuf))
	}
}
