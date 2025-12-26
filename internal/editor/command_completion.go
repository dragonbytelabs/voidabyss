package editor

import (
	"sort"
	"strings"
)

// getAllCommands returns all available commands for completion
func (e *Editor) getAllCommands() []string {
	return []string{
		"q", "q!",
		"w", "wq",
		"e",
		"bn", "bnext",
		"bp", "bprev", "bprevious",
		"bd", "bdelete", "bd!", "bdelete!",
		"ls", "buffers",
		"vsplit", "vs",
		"split", "sp",
		"close",
		"only",
		"Explore", "Ex",
		"reg", "registers",
		"macros",
		"noh", "nohlsearch",
		"fold",
		"foldopen", "fo",
		"foldclose", "fc",
		"foldall", "fca",
		"unfoldall", "ufa",
		"foldinfo",
		"colorscheme", "colorschemes",
		"set",
		"help",
	}
}

// getCommandCompletions returns commands that match the given prefix
func (e *Editor) getCommandCompletions(prefix string) []string {
	if prefix == "" {
		return nil
	}

	allCommands := e.getAllCommands()
	matches := make([]string, 0)

	for _, cmd := range allCommands {
		if strings.HasPrefix(cmd, prefix) {
			matches = append(matches, cmd)
		}
	}

	// Sort matches for consistent ordering
	sort.Strings(matches)

	return matches
}

// completeCommand cycles through available command completions
func (e *Editor) completeCommand(forward bool) {
	input := string(e.cmdBuf)

	// If not currently completing, find matches for current input
	if e.cmdCompletionIdx == -1 {
		// Save current input
		e.cmdCompletionSave = append([]rune(nil), e.cmdBuf...)

		// Get completions
		e.cmdCompletions = e.getCommandCompletions(input)

		if len(e.cmdCompletions) == 0 {
			// No matches, do nothing
			return
		}

		if forward {
			// Start at first completion
			e.cmdCompletionIdx = 0
		} else {
			// Backward from original: go to last completion
			e.cmdCompletionIdx = len(e.cmdCompletions) - 1
		}
		e.cmdBuf = []rune(e.cmdCompletions[e.cmdCompletionIdx])
		return
	}

	// Already completing, cycle through matches
	if len(e.cmdCompletions) == 0 {
		return
	}

	if forward {
		e.cmdCompletionIdx++
		if e.cmdCompletionIdx >= len(e.cmdCompletions) {
			// Wrap back to original input
			e.cmdBuf = append([]rune(nil), e.cmdCompletionSave...)
			e.cmdCompletionIdx = -1
			e.cmdCompletions = nil
			e.cmdCompletionSave = nil
			return
		}
	} else {
		// Backward (Shift+Tab)
		e.cmdCompletionIdx--
		if e.cmdCompletionIdx < 0 {
			// Wrap back to original input
			e.cmdBuf = append([]rune(nil), e.cmdCompletionSave...)
			e.cmdCompletionIdx = -1
			e.cmdCompletions = nil
			e.cmdCompletionSave = nil
			return
		}
	}

	e.cmdBuf = []rune(e.cmdCompletions[e.cmdCompletionIdx])
}

// resetCommandCompletion resets the completion state
func (e *Editor) resetCommandCompletion() {
	e.cmdCompletionIdx = -1
	e.cmdCompletions = nil
	e.cmdCompletionSave = nil
}
