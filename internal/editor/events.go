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
		e.loader.CallEventHandler( handler.Fn, eventData)
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
