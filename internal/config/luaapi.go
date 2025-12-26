package config

import (
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// SetupVbAPI sets up the complete vb table and functions in Lua
func (l *Loader) SetupVbAPI() {
	vbTable := l.L.NewTable()
	l.L.SetGlobal("vb", vbTable)

	// vb.version
	l.L.SetField(vbTable, "version", lua.LString(Version))

	// vb.opt (with metatable for assignment and methods)
	l.setupOptTable(vbTable)

	// vb.keymap(mode, lhs, rhs, opts)
	l.L.SetField(vbTable, "keymap", l.L.NewFunction(l.luaKeymap))

	// vb.command(name, rhs, opts)
	l.L.SetField(vbTable, "command", l.L.NewFunction(l.luaCommand))

	// vb.on(event, fn, opts)
	l.L.SetField(vbTable, "on", l.L.NewFunction(l.luaOn))

	// vb.buf (buffer operations)
	l.setupBufTable(vbTable)

	// vb.state (persistent storage)
	l.setupStateTable(vbTable)

	// vb.notify(msg, level)
	l.L.SetField(vbTable, "notify", l.L.NewFunction(l.luaNotify))

	// vb.has(feature)
	l.L.SetField(vbTable, "has", l.L.NewFunction(l.luaHas))

	// vb.plugins (for legacy compatibility)
	pluginsTable := l.L.NewTable()
	l.L.SetField(vbTable, "plugins", pluginsTable)
}

// setupOptTable creates vb.opt with metatable for property access
func (l *Loader) setupOptTable(vbTable *lua.LTable) {
	optTable := l.L.NewTable()

	// Add :get and :set methods
	getFunc := l.L.NewFunction(l.luaOptGet)
	setFunc := l.L.NewFunction(l.luaOptSet)

	l.L.SetField(optTable, "get", getFunc)
	l.L.SetField(optTable, "set", setFunc)

	// Create metatable for property assignment
	mt := l.L.NewTable()
	l.L.SetField(mt, "__index", l.L.NewFunction(l.luaOptIndex))
	l.L.SetField(mt, "__newindex", l.L.NewFunction(l.luaOptNewIndex))
	l.L.SetMetatable(optTable, mt)

	l.L.SetField(vbTable, "opt", optTable)
}

// luaOptGet implements vb.opt:get(key)
func (l *Loader) luaOptGet(L *lua.LState) int {
	key := L.CheckString(2) // Skip self (1)
	val := l.getOption(key)
	L.Push(val)
	return 1
}

// luaOptSet implements vb.opt:set(key, value)
func (l *Loader) luaOptSet(L *lua.LState) int {
	key := L.CheckString(2) // Skip self (1)
	value := L.CheckAny(3)
	l.setOption(key, value)
	return 0
}

// luaOptIndex implements property read (vb.opt.tabwidth)
func (l *Loader) luaOptIndex(L *lua.LState) int {
	key := L.CheckString(2)

	// Check if it's a method first
	if key == "get" || key == "set" {
		return 0
	}

	val := l.getOption(key)
	L.Push(val)
	return 1
}

// luaOptNewIndex implements property write (vb.opt.tabwidth = 4)
func (l *Loader) luaOptNewIndex(L *lua.LState) int {
	key := L.CheckString(2)
	value := L.CheckAny(3)
	l.setOption(key, value)
	return 0
}

// getOption retrieves an option value
func (l *Loader) getOption(key string) lua.LValue {
	opts := l.config.Options
	switch key {
	case "tabwidth":
		return lua.LNumber(opts.TabWidth)
	case "expandtab":
		return lua.LBool(opts.ExpandTab)
	case "cursorline":
		return lua.LBool(opts.CursorLine)
	case "relativenumber":
		return lua.LBool(opts.RelativeNumber)
	case "number":
		return lua.LBool(opts.Number)
	case "wrap":
		return lua.LBool(opts.Wrap)
	case "scrolloff":
		return lua.LNumber(opts.ScrollOff)
	case "leader":
		return lua.LString(opts.Leader)
	case "statusline":
		return lua.LString(opts.StatusLine)
	default:
		return lua.LNil
	}
}

// setOption sets an option value
func (l *Loader) setOption(key string, value lua.LValue) {
	opts := l.config.Options
	switch key {
	case "tabwidth":
		if num, ok := value.(lua.LNumber); ok {
			opts.TabWidth = int(num)
			l.config.TabWidth = int(num) // Legacy field
		}
	case "expandtab":
		if b, ok := value.(lua.LBool); ok {
			opts.ExpandTab = bool(b)
		}
	case "cursorline":
		if b, ok := value.(lua.LBool); ok {
			opts.CursorLine = bool(b)
		}
	case "relativenumber":
		if b, ok := value.(lua.LBool); ok {
			opts.RelativeNumber = bool(b)
			l.config.RelativeLineNums = bool(b) // Legacy field
		}
	case "number":
		if b, ok := value.(lua.LBool); ok {
			opts.Number = bool(b)
			l.config.ShowLineNumbers = bool(b) // Legacy field
		}
	case "wrap":
		if b, ok := value.(lua.LBool); ok {
			opts.Wrap = bool(b)
		}
	case "scrolloff":
		if num, ok := value.(lua.LNumber); ok {
			opts.ScrollOff = int(num)
		}
	case "leader":
		if str, ok := value.(lua.LString); ok {
			opts.Leader = string(str)
		}
	case "statusline":
		if str, ok := value.(lua.LString); ok {
			opts.StatusLine = string(str)
		}
	}
}

// luaKeymap implements vb.keymap(mode, lhs, rhs, opts)
func (l *Loader) luaKeymap(L *lua.LState) int {
	mode := L.CheckString(1)
	lhs := L.CheckString(2)
	rhs := L.Get(3) // Can be string or function

	opts := DefaultKeyMapOpts()
	if L.GetTop() >= 4 {
		if optsTable, ok := L.Get(4).(*lua.LTable); ok {
			if v := L.GetField(optsTable, "noremap"); v != lua.LNil {
				if b, ok := v.(lua.LBool); ok {
					opts.Noremap = bool(b)
				}
			}
			if v := L.GetField(optsTable, "silent"); v != lua.LNil {
				if b, ok := v.(lua.LBool); ok {
					opts.Silent = bool(b)
				}
			}
			if v := L.GetField(optsTable, "desc"); v != lua.LNil {
				if s, ok := v.(lua.LString); ok {
					opts.Desc = string(s)
				}
			}
		}
	}

	mapping := KeyMapping{
		Mode: mode,
		LHS:  lhs,
		Opts: opts,
	}

	// Check if rhs is a function or string
	if fn, ok := rhs.(*lua.LFunction); ok {
		mapping.Fn = fn
		mapping.IsFunc = true
	} else if str, ok := rhs.(lua.LString); ok {
		mapping.RHS = string(str)
		mapping.IsFunc = false
	}

	l.config.KeyMappings = append(l.config.KeyMappings, mapping)

	// Also store in legacy format for backwards compatibility
	if !mapping.IsFunc {
		key := mode + ":" + lhs
		l.config.KeyMaps[key] = KeyMap{
			Mode: mode,
			From: lhs,
			To:   mapping.RHS,
		}
	}

	return 0
}

// luaCommand implements vb.command(name, rhs, opts)
func (l *Loader) luaCommand(L *lua.LState) int {
	name := L.CheckString(1)
	rhs := L.Get(2) // Can be string or function

	opts := CommandOpts{NArgs: 0}
	if L.GetTop() >= 3 {
		if optsTable, ok := L.Get(3).(*lua.LTable); ok {
			if v := L.GetField(optsTable, "desc"); v != lua.LNil {
				if s, ok := v.(lua.LString); ok {
					opts.Desc = string(s)
				}
			}
			if v := L.GetField(optsTable, "nargs"); v != lua.LNil {
				if n, ok := v.(lua.LNumber); ok {
					opts.NArgs = int(n)
				} else if s, ok := v.(lua.LString); ok && string(s) == "?" {
					opts.NArgs = -1
				}
			}
		}
	}

	cmd := Command{
		Name: name,
		Opts: opts,
	}

	if fn, ok := rhs.(*lua.LFunction); ok {
		cmd.Fn = fn
		cmd.IsFunc = true
	} else if str, ok := rhs.(lua.LString); ok {
		cmd.RHS = string(str)
		cmd.IsFunc = false
	}

	l.config.Commands = append(l.config.Commands, cmd)
	return 0
}

// luaOn implements vb.on(event, fn, opts)
func (l *Loader) luaOn(L *lua.LState) int {
	event := L.CheckString(1)
	fn := L.CheckFunction(2)

	opts := EventOpts{}
	if L.GetTop() >= 3 {
		if optsTable, ok := L.Get(3).(*lua.LTable); ok {
			if v := L.GetField(optsTable, "pattern"); v != lua.LNil {
				if s, ok := v.(lua.LString); ok {
					opts.Pattern = string(s)
				}
			}
			if v := L.GetField(optsTable, "once"); v != lua.LNil {
				if b, ok := v.(lua.LBool); ok {
					opts.Once = bool(b)
				}
			}
		}
	}

	handler := EventHandler{
		Event: event,
		Fn:    fn,
		Opts:  opts,
	}

	l.config.EventHandlers = append(l.config.EventHandlers, handler)
	return 0
}

// setupBufTable creates vb.buf with buffer operations
func (l *Loader) setupBufTable(vbTable *lua.LTable) {
	bufTable := l.L.NewTable()

	// These will be stubs for now - actual implementation needs editor context
	l.L.SetField(bufTable, "get_text", l.L.NewFunction(l.luaBufGetText))
	l.L.SetField(bufTable, "set_text", l.L.NewFunction(l.luaBufSetText))
	l.L.SetField(bufTable, "get_name", l.L.NewFunction(l.luaBufGetName))
	l.L.SetField(bufTable, "set_name", l.L.NewFunction(l.luaBufSetName))
	l.L.SetField(bufTable, "cursor", l.L.NewFunction(l.luaBufCursor))
	l.L.SetField(bufTable, "set_cursor", l.L.NewFunction(l.luaBufSetCursor))
	l.L.SetField(bufTable, "line", l.L.NewFunction(l.luaBufLine))
	l.L.SetField(bufTable, "set_line", l.L.NewFunction(l.luaBufSetLine))
	l.L.SetField(bufTable, "insert", l.L.NewFunction(l.luaBufInsert))
	l.L.SetField(bufTable, "delete", l.L.NewFunction(l.luaBufDelete))
	l.L.SetField(bufTable, "len", l.L.NewFunction(l.luaBufLen))

	l.L.SetField(vbTable, "buf", bufTable)
}

// Buffer operation stubs (will need editor context for real implementation)
func (l *Loader) luaBufGetText(L *lua.LState) int {
	L.Push(lua.LString(""))
	return 1
}

func (l *Loader) luaBufSetText(L *lua.LState) int {
	return 0
}

func (l *Loader) luaBufGetName(L *lua.LState) int {
	L.Push(lua.LString(""))
	return 1
}

func (l *Loader) luaBufSetName(L *lua.LState) int {
	return 0
}

func (l *Loader) luaBufCursor(L *lua.LState) int {
	L.Push(lua.LNumber(0))
	L.Push(lua.LNumber(0))
	return 2
}

func (l *Loader) luaBufSetCursor(L *lua.LState) int {
	return 0
}

func (l *Loader) luaBufLine(L *lua.LState) int {
	L.Push(lua.LString(""))
	return 1
}

func (l *Loader) luaBufSetLine(L *lua.LState) int {
	return 0
}

func (l *Loader) luaBufInsert(L *lua.LState) int {
	return 0
}

func (l *Loader) luaBufDelete(L *lua.LState) int {
	return 0
}

func (l *Loader) luaBufLen(L *lua.LState) int {
	L.Push(lua.LNumber(0))
	return 1
}

// setupStateTable creates vb.state with persistent storage
func (l *Loader) setupStateTable(vbTable *lua.LTable) {
	stateTable := l.L.NewTable()

	l.L.SetField(stateTable, "get", l.L.NewFunction(l.luaStateGet))
	l.L.SetField(stateTable, "set", l.L.NewFunction(l.luaStateSet))
	l.L.SetField(stateTable, "del", l.L.NewFunction(l.luaStateDel))
	l.L.SetField(stateTable, "keys", l.L.NewFunction(l.luaStateKeys))

	l.L.SetField(vbTable, "state", stateTable)
}

// luaStateGet implements vb.state.get(key, default)
func (l *Loader) luaStateGet(L *lua.LState) int {
	key := L.CheckString(1)
	defaultVal := L.Get(2)

	val := l.config.State.Get(key, nil)
	if val == nil {
		L.Push(defaultVal)
		return 1
	}

	// Convert Go value to Lua value
	L.Push(goToLua(L, val))
	return 1
}

// luaStateSet implements vb.state.set(key, value)
func (l *Loader) luaStateSet(L *lua.LState) int {
	key := L.CheckString(1)
	value := L.CheckAny(2)

	// Convert Lua value to Go value
	goVal := luaToGo(value)
	l.config.State.Set(key, goVal)
	return 0
}

// luaStateDel implements vb.state.del(key)
func (l *Loader) luaStateDel(L *lua.LState) int {
	key := L.CheckString(1)
	l.config.State.Delete(key)
	return 0
}

// luaStateKeys implements vb.state.keys()
func (l *Loader) luaStateKeys(L *lua.LState) int {
	keys := l.config.State.Keys()
	tbl := L.NewTable()
	for i, key := range keys {
		L.RawSetInt(tbl, i+1, lua.LString(key))
	}
	L.Push(tbl)
	return 1
}

// luaNotify implements vb.notify(msg, level)
func (l *Loader) luaNotify(L *lua.LState) int {
	msg := L.CheckString(1)
	level := "info"
	if L.GetTop() >= 2 {
		level = L.CheckString(2)
	}

	// Store notification for later retrieval
	// For now, just format it for status display
	formatted := fmt.Sprintf("[%s] %s", strings.ToUpper(level), msg)
	_ = formatted // Will be used when integrated with editor

	return 0
}

// luaHas implements vb.has(feature)
func (l *Loader) luaHas(L *lua.LState) int {
	feature := L.CheckString(1)
	has := Features[feature]
	L.Push(lua.LBool(has))
	return 1
}

// Helper functions for Lua<->Go conversion
func luaToGo(value lua.LValue) interface{} {
	switch v := value.(type) {
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case lua.LBool:
		return bool(v)
	case *lua.LTable:
		// Try to determine if it's an array or map
		result := make(map[string]interface{})
		v.ForEach(func(key, val lua.LValue) {
			if keyStr, ok := key.(lua.LString); ok {
				result[string(keyStr)] = luaToGo(val)
			}
		})
		return result
	default:
		return nil
	}
}

func goToLua(L *lua.LState, value interface{}) lua.LValue {
	switch v := value.(type) {
	case string:
		return lua.LString(v)
	case float64:
		return lua.LNumber(v)
	case int:
		return lua.LNumber(v)
	case bool:
		return lua.LBool(v)
	case map[string]interface{}:
		tbl := L.NewTable()
		for key, val := range v {
			L.SetField(tbl, key, goToLua(L, val))
		}
		return tbl
	default:
		return lua.LNil
	}
}
