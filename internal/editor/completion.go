package editor

import (
	"sort"
	"strings"
)

// buildWordIndex extracts all words from the buffer, excluding the word at cursor
func (e *Editor) buildWordIndex(excludeStart, excludeEnd int) []string {
	text := e.buffer.String()
	words := make(map[string]bool)

	var current strings.Builder
	wordStart := 0
	pos := 0

	for _, r := range text {
		if isWordChar(r) {
			if current.Len() == 0 {
				wordStart = pos
			}
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				word := current.String()
				wordEnd := pos
				// Exclude word if it overlaps with the cursor range
				if len(word) > 1 && !(wordStart <= excludeEnd && wordEnd >= excludeStart) {
					words[word] = true
				}
				current.Reset()
			}
		}
		pos++
	}
	// last word
	if current.Len() > 0 {
		word := current.String()
		wordEnd := pos
		if len(word) > 1 && !(wordStart <= excludeEnd && wordEnd >= excludeStart) {
			words[word] = true
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(words))
	for word := range words {
		result = append(result, word)
	}

	// Sort for deterministic order: alphabetically
	sort.Strings(result)

	return result
}

// getCurrentWord returns the partial word before the cursor
func (e *Editor) getCurrentWord() (word string, startPos int) {
	pos := e.posFromCursor()
	if pos == 0 {
		return "", pos
	}

	// Get buffer as rune slice for proper Unicode handling
	runes := []rune(e.buffer.String())
	if pos > len(runes) {
		pos = len(runes)
	}

	// Walk backwards to find word start
	start := pos
	for start > 0 {
		r := runes[start-1]
		if !isWordChar(r) {
			break
		}
		start--
	}

	if start >= pos {
		return "", pos
	}

	return string(runes[start:pos]), start
}

// startCompletion initiates completion mode with initial index
func (e *Editor) startCompletion(initialIndex int) {
	prefix, startPos := e.getCurrentWord()
	if prefix == "" {
		e.statusMsg = "no word to complete"
		return
	}

	// Build word index, excluding the word at cursor
	cursorPos := e.posFromCursor()
	allWords := e.buildWordIndex(startPos, cursorPos)

	// Filter words that start with prefix
	var candidates []string
	for _, word := range allWords {
		if strings.HasPrefix(word, prefix) && word != prefix {
			candidates = append(candidates, word)
		}
	}

	if len(candidates) == 0 {
		e.statusMsg = "no completions found"
		return
	}

	// Sort candidates for stable, predictable order
	// Primary: length (shorter first - more common/simple completions)
	// Secondary: alphabetical (deterministic)
	sort.SliceStable(candidates, func(i, j int) bool {
		lenI, lenJ := len(candidates[i]), len(candidates[j])
		if lenI != lenJ {
			return lenI < lenJ
		}
		return candidates[i] < candidates[j]
	})

	e.completionActive = true
	e.completionCandidates = candidates
	// Set initial index (0 for forward/Ctrl-N, last for backward/Ctrl-P)
	if initialIndex < 0 {
		e.completionIndex = len(candidates) - 1
	} else {
		e.completionIndex = initialIndex
	}
	e.completionPrefix = prefix
	e.completionStartPos = startPos

	// Apply first completion
	e.applyCompletion()
}

// cycleCompletion moves to next/previous candidate
func (e *Editor) cycleCompletion(forward bool) {
	if !e.completionActive {
		// Start completion with appropriate initial index
		if forward {
			e.startCompletion(0) // Start at first candidate for Ctrl-N
		} else {
			e.startCompletion(-1) // Start at last candidate for Ctrl-P
		}
		return
	}

	// We're already in completion mode, so cycle the index
	if forward {
		e.completionIndex++
		if e.completionIndex >= len(e.completionCandidates) {
			e.completionIndex = 0
		}
	} else {
		e.completionIndex--
		if e.completionIndex < 0 {
			e.completionIndex = len(e.completionCandidates) - 1
		}
	}

	// Apply the newly selected completion
	e.applyCompletion()
}

// applyCompletion replaces the current word with the selected candidate
func (e *Editor) applyCompletion() {
	if !e.completionActive || len(e.completionCandidates) == 0 {
		return
	}

	candidate := e.completionCandidates[e.completionIndex]

	// Delete from start of completion to current cursor
	// (which includes any previously applied completion)
	currentPos := e.posFromCursor()
	deleteStart := e.completionStartPos

	if currentPos > deleteStart {
		_ = e.buffer.Delete(deleteStart, currentPos)
	}

	// Insert the new completion
	_ = e.buffer.Insert(deleteStart, candidate)
	e.setCursorFromPos(deleteStart + len(candidate))
	e.wantX = e.cx

	// Update status with completion info
	e.statusMsg = ""
	e.updateCompletionPopup()
}

// updateCompletionPopup shows the completion candidates
func (e *Editor) updateCompletionPopup() {
	if !e.completionActive {
		e.popupActive = false
		return
	}

	lines := make([]string, 0, len(e.completionCandidates))
	for i, candidate := range e.completionCandidates {
		prefix := "  "
		if i == e.completionIndex {
			prefix = "> "
		}

		// Highlight matching prefix by surrounding it with brackets
		// e.g., if prefix="hel" and candidate="hello", show "[hel]lo"
		matchLen := len(e.completionPrefix)
		if matchLen > 0 && matchLen <= len(candidate) {
			highlighted := "[" + candidate[:matchLen] + "]" + candidate[matchLen:]
			lines = append(lines, prefix+highlighted)
		} else {
			lines = append(lines, prefix+candidate)
		}
	}

	e.popupActive = true
	e.popupTitle = "Completions"
	e.popupLines = lines
	e.popupScroll = 0

	// Auto-scroll to show current selection
	if e.completionIndex > 5 {
		e.popupScroll = e.completionIndex - 5
	}
}

// cancelCompletion exits completion mode
func (e *Editor) cancelCompletion() {
	e.completionActive = false
	e.completionCandidates = nil
	e.completionIndex = 0
	e.completionPrefix = ""
	e.popupActive = false
	e.statusMsg = ""
}
