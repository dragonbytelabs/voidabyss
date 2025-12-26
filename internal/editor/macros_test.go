package editor

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestMacroRecordAndPlayback(t *testing.T) {
	e := newTestEditor(t, "line1\nline2\nline3")

	// Start recording to register 'a'
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))

	t.Logf("After 'q': statusMsg=%q (len=%d), awaitingRegister=%v", e.statusMsg, len(e.statusMsg), e.awaitingRegister)
	if e.statusMsg != "record macro to register: " {
		t.Fatalf("Unexpected statusMsg after 'q': got %q, want %q", e.statusMsg, "record macro to register: ")
	}

	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))

	t.Logf("After 'a': statusMsg=%q, recordingMacro=%v, recordingRegister=%c",
		e.statusMsg, e.recordingMacro, e.recordingRegister)

	if !e.recordingMacro {
		t.Fatal("should be recording")
	}
	if e.recordingRegister != 'a' {
		t.Errorf("recording register = %c, want 'a'", e.recordingRegister)
	}

	// Record: I#<Esc>j (add # to start of line, move down)
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'I', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '#', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone))

	// Stop recording
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))

	if e.recordingMacro {
		t.Fatal("should have stopped recording")
	}

	// Verify first line has #
	expected := "#line1\nline2\nline3"
	if got := e.buffer.String(); got != expected {
		t.Errorf("after recording:\nexpected: %q\ngot: %q", expected, got)
	}

	// Cursor should be on line 2
	if e.cy != 1 {
		t.Errorf("cy = %d, want 1", e.cy)
	}

	// Playback macro
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '@', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))

	expected = "#line1\n#line2\nline3"
	if got := e.buffer.String(); got != expected {
		t.Errorf("after playback:\nexpected: %q\ngot: %q", expected, got)
	}

	// Cursor should be on line 3
	if e.cy != 2 {
		t.Errorf("cy = %d, want 2", e.cy)
	}
}

func TestMacroPlaybackCount(t *testing.T) {
	e := newTestEditor(t, "a\nb\nc\nd\ne")

	// Record: x (delete character)
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))

	// Play 3 times
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '@', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))

	expected := "\n\n\n\ne"
	if got := e.buffer.String(); got != expected {
		t.Errorf("after 3x playback:\nexpected: %q\ngot: %q", expected, got)
	}

	// Cursor should be on line 4 (recorded once during qa...q, then played 3 more times = 4 total)
	if e.cy != 4 {
		t.Errorf("cy = %d, want 4", e.cy)
	}
}

func TestMacroEmpty(t *testing.T) {
	e := newTestEditor(t, "test")

	// Try to play non-existent macro
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '@', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone))

	if e.statusMsg != "no macro in register @z" {
		t.Errorf("statusMsg = %q, want error message", e.statusMsg)
	}

	// Buffer should be unchanged
	if got := e.buffer.String(); got != "test" {
		t.Errorf("buffer = %q, want 'test'", got)
	}
}

func TestMacroComplexSequence(t *testing.T) {
	e := newTestEditor(t, "apple\nbanana\ncherry")

	// Record: A!<Esc>j (append ! to end of line, move down)
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'A', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '!', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))

	expected := "apple!\nbanana\ncherry"
	if got := e.buffer.String(); got != expected {
		t.Errorf("after recording:\nexpected: %q\ngot: %q", expected, got)
	}

	// Play twice
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '2', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, '@', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))

	expected = "apple!\nbanana!\ncherry!"
	if got := e.buffer.String(); got != expected {
		t.Errorf("after playback:\nexpected: %q\ngot: %q", expected, got)
	}
}

func TestMacroDoesNotRecordDuringPlayback(t *testing.T) {
	e := newTestEditor(t, "test")

	// Record macro 'a' that inserts "x"
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))

	// Start recording macro 'b'
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'b', tcell.ModNone))

	oldRecording := e.recordingMacro
	macroKeysLen := len(e.macroKeys)

	// Play macro 'a' (should not be recorded in 'b')
	e.playbackMacro('a', 1)

	// Recording should still be active
	if e.recordingMacro != oldRecording {
		t.Error("recording state should not change during playback")
	}

	// Stop recording 'b'
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))

	// Macro 'b' should only have the keys explicitly typed, not from playback
	if len(e.macros['b'].keys) > macroKeysLen {
		t.Error("macro 'b' should not contain keys from played macro 'a'")
	}
}

func TestMacroStopRecordingWithQ(t *testing.T) {
	e := newTestEditor(t, "test")

	// Start recording
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))

	if !e.recordingMacro {
		t.Fatal("should be recording")
	}

	// Type some commands
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'h', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone))

	// Stop with 'q'
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))

	if e.recordingMacro {
		t.Fatal("should have stopped recording")
	}

	// Macro should exist
	if _, exists := e.macros['a']; !exists {
		t.Error("macro 'a' should exist")
	}
}

func TestMacroFormatDisplay(t *testing.T) {
	e := newTestEditor(t, "test")

	// Record a simple macro
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))
	e.handleKey(tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone))

	lines := e.formatMacros()
	if len(lines) == 0 {
		t.Fatal("should have formatted macro lines")
	}

	// Should show register 'a'
	found := false
	for _, line := range lines {
		if len(line) > 0 && line[0] == 'a' {
			found = true
			// Should show keys
			if !contains(line, "dd") {
				t.Errorf("line should contain 'dd', got: %q", line)
			}
			break
		}
	}

	if !found {
		t.Error("should have found macro 'a' in formatted output")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
