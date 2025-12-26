package editor

// Completion control methods for Lua ctx

// CompleteNext advances to the next completion candidate
func (e *Editor) CompleteNext() {
	if !e.completionActive {
		e.startCompletion(0)
		return
	}

	if len(e.completionCandidates) == 0 {
		return
	}

	// Move to next candidate
	e.completionIndex = (e.completionIndex + 1) % len(e.completionCandidates)
	e.applyCompletion()
	e.updateCompletionPopup()
}

// CompletePrev moves to the previous completion candidate
func (e *Editor) CompletePrev() {
	if !e.completionActive {
		e.startCompletion(len(e.completionCandidates) - 1)
		return
	}

	if len(e.completionCandidates) == 0 {
		return
	}

	// Move to previous candidate
	e.completionIndex--
	if e.completionIndex < 0 {
		e.completionIndex = len(e.completionCandidates) - 1
	}
	e.applyCompletion()
	e.updateCompletionPopup()
}

// CompleteCancel cancels active completion
func (e *Editor) CompleteCancel() {
	if e.completionActive {
		e.cancelCompletion()
	}
}

// ExecCommand executes a command and returns true if editor should quit
func (e *Editor) ExecCommand(cmd string) bool {
	return e.exec(cmd)
}

// Feedkeys simulates key input (stub for now - would need key parsing from task 4)
func (e *Editor) Feedkeys(keys string) {
	// TODO: Implement once we have key notation parser (task 4)
	// For now, just set status message
	e.statusMsg = "feedkeys not yet implemented: " + keys
}
