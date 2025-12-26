package editor

import (
	"github.com/dragonbytelabs/voidabyss/internal/config"
)

// FireEvent triggers all event handlers registered for the given event name
func (e *Editor) FireEvent(eventName string, eventData map[string]interface{}) {
	if e.loader == nil {
		return
	}

	// Get all event handlers for this event
	handlers := e.getEventHandlers(eventName)
	if len(handlers) == 0 {
		return
	}

	// Call each handler
	for _, handler := range handlers {
		// Check if this is a once-only handler
		if handler.Opts.Once {
			// Remove it after this call
			defer e.removeEventHandler(eventName, &handler)
		}

		// Check pattern matching if specified
		if handler.Opts.Pattern != "" {
			// For now, simple string comparison with filename
			// TODO: Implement glob matching
			if e.filename != handler.Opts.Pattern {
				continue
			}
		}

		// Call the Lua handler
		e.loader.CallEventHandler(handler.Fn, eventData)
	}
}

// getEventHandlers returns all handlers registered for an event
func (e *Editor) getEventHandlers(eventName string) []config.EventHandler {
	if e.config == nil {
		return nil
	}

	var handlers []config.EventHandler
	for _, h := range e.config.EventHandlers {
		if h.Event == eventName {
			handlers = append(handlers, h)
		}
	}

	return handlers
}

// removeEventHandler removes a once-only event handler
func (e *Editor) removeEventHandler(eventName string, handler *config.EventHandler) {
	if e.config == nil {
		return
	}

	// Find and remove the handler
	for i, h := range e.config.EventHandlers {
		if h.Event == eventName && h.Fn == handler.Fn {
			// Remove by replacing with last element and truncating
			e.config.EventHandlers[i] = e.config.EventHandlers[len(e.config.EventHandlers)-1]
			e.config.EventHandlers = e.config.EventHandlers[:len(e.config.EventHandlers)-1]
			break
		}
	}
}

// Event helper methods for common events

// FireEditorReady fires when the editor finishes initialization
func (e *Editor) FireEditorReady() {
	e.FireEvent("EditorReady", map[string]interface{}{
		"version": config.Version,
	})
}

// FireBufRead fires when a buffer is loaded
func (e *Editor) FireBufRead() {
	e.FireEvent("BufRead", map[string]interface{}{
		"file": e.filename,
	})
}

// FireBufWritePre fires before saving a buffer
func (e *Editor) FireBufWritePre() {
	e.FireEvent("BufWritePre", map[string]interface{}{
		"file": e.filename,
	})
}

// FireBufWritePost fires after saving a buffer
func (e *Editor) FireBufWritePost() {
	e.FireEvent("BufWritePost", map[string]interface{}{
		"file": e.filename,
	})
}

// FireModeChanged fires when the editor mode changes
func (e *Editor) FireModeChanged(oldMode, newMode Mode) {
	e.FireEvent("ModeChanged", map[string]interface{}{
		"old_mode": oldMode.string(),
		"new_mode": newMode.string(),
	})
}

// FireInsertEnter fires when entering insert mode
func (e *Editor) FireInsertEnter() {
	e.FireEvent("InsertEnter", map[string]interface{}{})
}

// FireInsertLeave fires when leaving insert mode
func (e *Editor) FireInsertLeave() {
	e.FireEvent("InsertLeave", map[string]interface{}{})
}

// FireBufEnter fires when switching to a buffer
func (e *Editor) FireBufEnter() {
	e.FireEvent("BufEnter", map[string]interface{}{
		"file":  e.filename,
		"bufnr": e.currentBuffer,
	})
}

// FireBufLeave fires when leaving a buffer
func (e *Editor) FireBufLeave() {
	e.FireEvent("BufLeave", map[string]interface{}{
		"file":  e.filename,
		"bufnr": e.currentBuffer,
	})
}

// FireBufNew fires when a new buffer is created
func (e *Editor) FireBufNew() {
	e.FireEvent("BufNew", map[string]interface{}{
		"file":  e.filename,
		"bufnr": len(e.buffers) - 1,
	})
}

// FireBufDelete fires before a buffer is deleted
func (e *Editor) FireBufDelete(bufIndex int) {
	filename := ""
	if bufIndex >= 0 && bufIndex < len(e.buffers) {
		filename = e.buffers[bufIndex].filename
	}
	e.FireEvent("BufDelete", map[string]interface{}{
		"file":  filename,
		"bufnr": bufIndex,
	})
}

// FireTextChanged fires when text is modified
func (e *Editor) FireTextChanged() {
	e.FireEvent("TextChanged", map[string]interface{}{
		"file": e.filename,
	})
}

// FireTextChangedI fires when text is modified in insert mode
func (e *Editor) FireTextChangedI() {
	e.FireEvent("TextChangedI", map[string]interface{}{
		"file": e.filename,
	})
}

// FireCursorMoved fires when cursor moves in normal mode
func (e *Editor) FireCursorMoved() {
	e.FireEvent("CursorMoved", map[string]interface{}{
		"line": e.cy,
		"col":  e.cx,
	})
}

// FireCursorMovedI fires when cursor moves in insert mode
func (e *Editor) FireCursorMovedI() {
	e.FireEvent("CursorMovedI", map[string]interface{}{
		"line": e.cy,
		"col":  e.cx,
	})
}

// FireVimEnter fires when editor starts
func (e *Editor) FireVimEnter() {
	e.FireEvent("VimEnter", map[string]interface{}{})
}

// FireVimLeave fires before editor exits
func (e *Editor) FireVimLeave() {
	e.FireEvent("VimLeave", map[string]interface{}{})
}

// FireFileType fires when filetype is detected
func (e *Editor) FireFileType(filetype string) {
	e.FireEvent("FileType", map[string]interface{}{
		"filetype": filetype,
		"file":     e.filename,
	})
}

// FireVisualEnter fires when entering visual mode
func (e *Editor) FireVisualEnter() {
	e.FireEvent("VisualEnter", map[string]interface{}{})
}

// FireVisualLeave fires when leaving visual mode
func (e *Editor) FireVisualLeave() {
	e.FireEvent("VisualLeave", map[string]interface{}{})
}

// FireSearchComplete fires when search completes
func (e *Editor) FireSearchComplete(query string, matchCount int) {
	e.FireEvent("SearchComplete", map[string]interface{}{
		"query":   query,
		"matches": matchCount,
	})
}

// setModeWithEvent sets the mode and fires ModeChanged event
func (e *Editor) setModeWithEvent(newMode Mode) {
	if e.mode == newMode {
		return
	}

	oldMode := e.mode
	e.mode = newMode

	// Fire mode-specific events
	if oldMode == ModeInsert {
		e.FireInsertLeave()
	}
	if newMode == ModeInsert {
		e.FireInsertEnter()
	}

	// Fire generic mode changed event
	e.FireModeChanged(oldMode, newMode)
}
