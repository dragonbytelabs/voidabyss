package config

import lua "github.com/yuin/gopher-lua"

// RegisterEventAPI adds event-related functions to the Lua state
func RegisterEventAPI(L *lua.LState) {
	// Create events table
	eventsTable := L.NewTable()

	// on_event(event_name, callback)
	L.SetField(eventsTable, "on", L.NewFunction(luaOnEvent))

	// trigger_event(event_name, data)
	L.SetField(eventsTable, "trigger", L.NewFunction(luaTriggerEvent))

	// Event constants
	eventsTable.RawSetString("BUFFER_OPENED", lua.LString(EventBufferOpened))
	eventsTable.RawSetString("BUFFER_CLOSED", lua.LString(EventBufferClosed))
	eventsTable.RawSetString("BUFFER_CHANGED", lua.LString(EventBufferChanged))
	eventsTable.RawSetString("BUFFER_SAVED", lua.LString(EventBufferSaved))
	eventsTable.RawSetString("CURSOR_MOVED", lua.LString(EventCursorMoved))

	L.SetGlobal("events", eventsTable)
}

// luaOnEvent registers an event callback from Lua
// Usage: events.on(events.BUFFER_OPENED, function(data) ... end)
func luaOnEvent(L *lua.LState) int {
	eventName := L.CheckString(1)
	callback := L.CheckFunction(2)

	event := EditorEvent(eventName)

	// Wrap the Lua function in a Go callback
	RegisterEventHook(event, func(data EventData) {
		// Create a new Lua state for this callback to avoid concurrency issues
		// Or we could use a channel to marshal back to main Lua state
		dataTable := L.NewTable()
		dataTable.RawSetString("event", lua.LString(data.Event))
		dataTable.RawSetString("filepath", lua.LString(data.FilePath))
		dataTable.RawSetString("line", lua.LNumber(data.Line))
		dataTable.RawSetString("col", lua.LNumber(data.Col))
		dataTable.RawSetString("content", lua.LString(data.Content))

		// Call the Lua callback
		L.Push(callback)
		L.Push(dataTable)
		if err := L.PCall(1, 0, nil); err != nil {
			// Log error but don't crash
			L.RaiseError("Event callback error: %v", err)
		}
	})

	return 0
}

// luaTriggerEvent fires an event from Lua
// Usage: events.trigger(events.BUFFER_OPENED, {filepath = "...", line = 1, col = 0})
func luaTriggerEvent(L *lua.LState) int {
	eventName := L.CheckString(1)
	dataTable := L.CheckTable(2)

	data := EventData{
		Event: EditorEvent(eventName),
	}

	if v := dataTable.RawGetString("filepath"); v != lua.LNil {
		data.FilePath = v.String()
	}
	if v := dataTable.RawGetString("line"); v != lua.LNil {
		data.Line = int(v.(lua.LNumber))
	}
	if v := dataTable.RawGetString("col"); v != lua.LNil {
		data.Col = int(v.(lua.LNumber))
	}
	if v := dataTable.RawGetString("content"); v != lua.LNil {
		data.Content = v.String()
	}

	TriggerEvent(data)
	return 0
}
