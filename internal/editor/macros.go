package editor

import "github.com/gdamore/tcell/v2"

// Macro represents a recorded sequence of key events
type Macro struct {
	keys []MacroKey
}

// MacroKey represents a recorded key event
type MacroKey struct {
	key tcell.Key
	ch  rune
	mod tcell.ModMask
}

// startRecording starts recording a macro into the specified register
func (e *Editor) startRecording(register rune) {
	if register < 'a' || register > 'z' {
		e.statusMsg = "macro register must be a-z"
		return
	}

	// If already recording, stop
	if e.recordingMacro {
		e.stopRecording()
		return
	}

	e.recordingMacro = true
	e.recordingRegister = register
	e.macroKeys = nil
	e.statusMsg = "recording @" + string(register)
}

// stopRecording stops recording the current macro
func (e *Editor) stopRecording() {
	if !e.recordingMacro {
		return
	}

	// Save the macro to the register
	if e.recordingRegister >= 'a' && e.recordingRegister <= 'z' {
		e.macros[e.recordingRegister] = Macro{keys: e.macroKeys}
	}

	e.recordingMacro = false
	e.statusMsg = "macro saved to @" + string(e.recordingRegister)
	e.recordingRegister = 0
	e.macroKeys = nil
}

// recordKey records a key event during macro recording
func (e *Editor) recordKey(k *tcell.EventKey) {
	if !e.recordingMacro {
		return
	}

	e.macroKeys = append(e.macroKeys, MacroKey{
		key: k.Key(),
		ch:  k.Rune(),
		mod: k.Modifiers(),
	})
}

// playbackMacro plays back a recorded macro
func (e *Editor) playbackMacro(register rune, count int) {
	if register < 'a' || register > 'z' {
		e.statusMsg = "macro register must be a-z"
		return
	}

	macro, exists := e.macros[register]
	if !exists || len(macro.keys) == 0 {
		e.statusMsg = "no macro in register @" + string(register)
		return
	}

	// Prevent recursive macro playback
	if e.playingMacro {
		e.statusMsg = "already playing macro"
		return
	}

	if count < 1 {
		count = 1
	}

	e.playingMacro = true
	defer func() { e.playingMacro = false }()

	// Play the macro 'count' times
	for i := 0; i < count; i++ {
		for _, mk := range macro.keys {
			event := tcell.NewEventKey(mk.key, mk.ch, mk.mod)

			// Don't record keys during playback
			oldRecording := e.recordingMacro
			e.recordingMacro = false

			e.handleKey(event)

			e.recordingMacro = oldRecording
		}
	}

	e.statusMsg = "macro @" + string(register) + " complete"
}

// formatMacros returns a formatted list of recorded macros for display
func (e *Editor) formatMacros() []string {
	lines := make([]string, 0, 26)

	for ch := 'a'; ch <= 'z'; ch++ {
		macro, exists := e.macros[ch]
		if !exists || len(macro.keys) == 0 {
			continue
		}

		// Format: "a  <keys> (5 keys)"
		preview := e.formatMacroKeys(macro.keys, 50)
		lines = append(lines, string(ch)+"  "+preview)
	}

	if len(lines) == 0 {
		lines = append(lines, "No macros recorded")
	}

	return lines
}

// formatMacroKeys creates a readable representation of macro keys
func (e *Editor) formatMacroKeys(keys []MacroKey, maxLen int) string {
	result := ""
	for _, mk := range keys {
		var keyStr string

		switch mk.key {
		case tcell.KeyRune:
			keyStr = string(mk.ch)
		case tcell.KeyEscape:
			keyStr = "<Esc>"
		case tcell.KeyEnter:
			keyStr = "<CR>"
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			keyStr = "<BS>"
		case tcell.KeyTab:
			keyStr = "<Tab>"
		case tcell.KeyDelete:
			keyStr = "<Del>"
		case tcell.KeyUp:
			keyStr = "<Up>"
		case tcell.KeyDown:
			keyStr = "<Down>"
		case tcell.KeyLeft:
			keyStr = "<Left>"
		case tcell.KeyRight:
			keyStr = "<Right>"
		case tcell.KeyCtrlR:
			keyStr = "<C-R>"
		case tcell.KeyCtrlN:
			keyStr = "<C-N>"
		case tcell.KeyCtrlP:
			keyStr = "<C-P>"
		case tcell.KeyCtrlW:
			keyStr = "<C-W>"
		case tcell.KeyCtrlO:
			keyStr = "<C-O>"
		case tcell.KeyCtrlI:
			keyStr = "<C-I>"
		default:
			keyStr = "<Key>"
		}

		if len(result)+len(keyStr) > maxLen {
			result += "..."
			break
		}
		result += keyStr
	}

	return result + " (" + string(rune('0'+len(keys))) + " keys)"
}
