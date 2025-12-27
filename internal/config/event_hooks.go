package config

import (
	"sync"
)

// EditorEvent represents different events that can occur in the editor
type EditorEvent string

const (
	EventBufferOpened  EditorEvent = "BufferOpened"
	EventBufferClosed  EditorEvent = "BufferClosed"
	EventBufferChanged EditorEvent = "BufferChanged"
	EventBufferSaved   EditorEvent = "BufferSaved"
	EventCursorMoved   EditorEvent = "CursorMoved"
)

// EventData contains information about an event
type EventData struct {
	Event    EditorEvent
	FilePath string
	Line     int
	Col      int
	Content  string
}

// EventCallback is a function that handles editor events
type EventCallback func(data EventData)

// eventHooks stores registered event callbacks
var (
	eventHooks = make(map[EditorEvent][]EventCallback)
	hooksMutex sync.RWMutex
)

// RegisterEventHook registers a callback for a specific editor event
func RegisterEventHook(event EditorEvent, callback EventCallback) {
	hooksMutex.Lock()
	defer hooksMutex.Unlock()
	eventHooks[event] = append(eventHooks[event], callback)
}

// TriggerEvent fires all callbacks registered for an event
func TriggerEvent(data EventData) {
	hooksMutex.RLock()
	callbacks := eventHooks[data.Event]
	hooksMutex.RUnlock()

	// Run callbacks in goroutines so they don't block the editor
	for _, callback := range callbacks {
		go callback(data)
	}
}

// ClearEventHooks removes all registered event hooks (useful for testing/reloading)
func ClearEventHooks() {
	hooksMutex.Lock()
	defer hooksMutex.Unlock()
	eventHooks = make(map[EditorEvent][]EventCallback)
}
