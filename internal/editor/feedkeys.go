package editor

import (
	"strings"

	"github.com/dragonbytelabs/voidabyss/internal/config"
	"github.com/gdamore/tcell/v2"
)

// Feedkeys executes a sequence of keys as if they were typed
// This is the execution engine for string RHS mappings
func (e *Editor) Feedkeys(keys string) {
	// Parse the key sequence with leader expansion
	leader := "\\" // Default leader
	if e.config != nil && e.config.Options != nil {
		leader = e.config.Options.Leader
	}

	parsedKeys := config.ParseKeyNotation(keys, leader)

	// Execute each key in sequence
	for _, ch := range parsedKeys {
		e.handleFeedkey(ch)
	}
}

// handleFeedkey processes a single character from feedkeys
func (e *Editor) handleFeedkey(ch rune) {
	// Check if this starts a command
	if ch == ':' && e.mode == ModeNormal {
		// Enter command mode
		e.mode = ModeCommand
		e.cmdBuf = []rune{':'}
		return
	}

	// Check if this starts a search
	if (ch == '/' || ch == '?') && e.mode == ModeNormal {
		// Enter search mode - use the search handling
		e.searchBuf = []rune{ch}
		e.searchForward = (ch == '/')
		return
	}

	// Handle based on mode
	switch e.mode {
	case ModeNormal:
		e.handleNormalFeedkey(ch)
	case ModeInsert:
		e.handleInsertFeedkey(ch)
	case ModeVisual:
		e.handleVisualFeedkey(ch)
	case ModeCommand:
		e.handleCommandFeedkey(ch)
	}
}

// handleNormalFeedkey processes a key in normal mode
func (e *Editor) handleNormalFeedkey(ch rune) {
	switch ch {
	case '\r': // Enter/CR
		// Move to start of next line
		e.moveDown(1)
		e.cx = 0
		e.wantX = 0

	case '\x1b': // Escape
		// Clear any pending operations
		e.pendingOp = 0
		e.pendingCount = 0
		e.statusMsg = ""

	case '\t': // Tab
		// Tab in normal mode - could be used for navigation
		// For now, do nothing

	default:
		// Create a synthetic event for normal key handling
		ev := tcell.NewEventKey(tcell.KeyRune, ch, tcell.ModNone)
		e.handleNormal(ev)
	}
}

// handleInsertFeedkey processes a key in insert mode
func (e *Editor) handleInsertFeedkey(ch rune) {
	switch ch {
	case '\r': // Enter/CR
		e.newline()

	case '\x1b': // Escape
		// Exit insert mode
		if e.mode == ModeInsert {
			// End undo group for this insert session
			e.buffer.EndUndoGroup()
			// Store last action for repeat
			e.last.insertText = e.insertCapture
			e.insertCapture = nil
		}
		e.mode = ModeNormal
		// Move cursor left (unless at start of line)
		if e.cx > 0 {
			e.cx--
		}
		e.wantX = e.cx

	case '\t': // Tab
		e.insertRune('\t')

	case '\x08', '\x7f': // Backspace or Delete
		e.backspace()

	default:
		e.insertRune(ch)
		if e.mode == ModeInsert {
			e.insertCapture = append(e.insertCapture, ch)
		}
	}
}

// handleVisualFeedkey processes a key in visual mode
func (e *Editor) handleVisualFeedkey(ch rune) {
	switch ch {
	case '\x1b': // Escape
		e.visualActive = false
		e.mode = ModeNormal
		e.statusMsg = ""

	case '\r': // Enter - for now just exit visual mode
		e.visualActive = false
		e.mode = ModeNormal

	default:
		// Create a synthetic event for visual key handling
		ev := tcell.NewEventKey(tcell.KeyRune, ch, tcell.ModNone)
		e.handleVisual(ev)
	}
}

// handleCommandFeedkey processes a key in command mode
func (e *Editor) handleCommandFeedkey(ch rune) {
	switch ch {
	case '\r': // Enter - execute command
		if len(e.cmdBuf) > 0 {
			// Remove leading ':'
			cmd := string(e.cmdBuf)
			cmd = strings.TrimPrefix(cmd, ":")
			e.exec(cmd)
		}
		e.mode = ModeNormal
		e.cmdBuf = nil

	case '\x1b': // Escape - cancel command
		e.mode = ModeNormal
		e.cmdBuf = nil
		e.statusMsg = ""

	case '\x08', '\x7f': // Backspace
		if len(e.cmdBuf) > 1 {
			e.cmdBuf = e.cmdBuf[:len(e.cmdBuf)-1]
		} else {
			// Cancel command mode if we delete the ':'
			e.mode = ModeNormal
			e.cmdBuf = nil
		}

	default:
		e.cmdBuf = append(e.cmdBuf, ch)
	}
}
