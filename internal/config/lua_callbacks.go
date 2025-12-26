package config

import (
	lua "github.com/yuin/gopher-lua"
)

// CallKeymapFunction calls a keymap Lua function with ctx
func (l *Loader) CallKeymapFunction(fn interface{}) error {
	if l.L == nil {
		return nil
	}

	luaFn, ok := fn.(*lua.LFunction)
	if !ok {
		return nil
	}

	// Create ctx userdata
	ctx := l.NewLuaContext()

	// Call function with ctx
	l.L.Push(luaFn)
	l.L.Push(ctx)

	if err := l.L.PCall(1, 0, nil); err != nil {
		l.Notifications.Push("Keymap error: "+err.Error(), NotifyError)
		return err
	}

	return nil
}

// CallCommandFunction calls a command Lua function with args and ctx
func (l *Loader) CallCommandFunction(fn interface{}, args string) error {
	if l.L == nil {
		return nil
	}

	luaFn, ok := fn.(*lua.LFunction)
	if !ok {
		return nil
	}

	// Create ctx userdata
	ctx := l.NewLuaContext()

	// Call function with (args, ctx)
	l.L.Push(luaFn)
	l.L.Push(lua.LString(args))
	l.L.Push(ctx)

	if err := l.L.PCall(2, 0, nil); err != nil {
		l.Notifications.Push("Command error: "+err.Error(), NotifyError)
		return err
	}

	return nil
}

// CallEventHandler calls an event handler Lua function with ctx
func (l *Loader) CallEventHandler(fn interface{}, eventData map[string]interface{}) error {
	if l.L == nil {
		return nil
	}

	luaFn, ok := fn.(*lua.LFunction)
	if !ok {
		return nil
	}

	// Create ctx userdata
	ctx := l.NewLuaContext()

	// Add event data to ctx (as fields in a table passed as second arg)
	dataTable := l.L.NewTable()
	for k, v := range eventData {
		l.L.SetField(dataTable, k, goToLua(l.L, v))
	}

	// Call function with (ctx, data)
	l.L.Push(luaFn)
	l.L.Push(ctx)
	l.L.Push(dataTable)

	if err := l.L.PCall(2, 0, nil); err != nil {
		l.Notifications.Push("Event handler error: "+err.Error(), NotifyError)
		return err
	}

	return nil
}
