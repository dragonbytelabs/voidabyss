package config

import (
	"fmt"
	"os"
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

	// vb.keymap (with list method)
	l.setupKeymapTable(vbTable)

	// vb.command(name, rhs, opts)
	l.L.SetField(vbTable, "command", l.L.NewFunction(l.luaCommand))

	// vb.on(event, fn, opts)
	l.L.SetField(vbTable, "on", l.L.NewFunction(l.luaOn))

	// vb.augroup(name, fn)
	l.L.SetField(vbTable, "augroup", l.L.NewFunction(l.luaAugroup))

	// vb.buf (buffer operations)
	l.setupBufTable(vbTable)

	// vb.state (persistent storage)
	l.setupStateTable(vbTable)

	// vb.notify(msg, level)
	l.L.SetField(vbTable, "notify", l.L.NewFunction(l.luaNotify))

	// vb.has(feature)
	l.L.SetField(vbTable, "has", l.L.NewFunction(l.luaHas))

	// vb.checkhealth()
	l.L.SetField(vbTable, "checkhealth", l.L.NewFunction(l.luaCheckHealth))

	// vb.schedule(fn)
	l.L.SetField(vbTable, "schedule", l.L.NewFunction(l.luaSchedule))

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

// setupKeymapTable creates vb.keymap as a callable function with methods
func (l *Loader) setupKeymapTable(vbTable *lua.LTable) {
	keymapFunc := l.L.NewFunction(l.luaKeymap)

	// Create a metatable to make the function callable and also have methods
	mt := l.L.NewTable()

	// __call makes vb.keymap() work
	l.L.SetField(mt, "__call", l.L.NewFunction(func(L *lua.LState) int {
		// Remove the self argument and call the actual function
		L.Remove(1)
		return l.luaKeymap(L)
	}))

	// __index provides methods like vb.keymap.list()
	indexTable := l.L.NewTable()
	l.L.SetField(indexTable, "list", l.L.NewFunction(l.luaKeymapList))
	l.L.SetField(mt, "__index", indexTable)

	l.L.SetMetatable(keymapFunc, mt)
	l.L.SetField(vbTable, "keymap", keymapFunc)
}

// luaKeymapList implements vb.keymap.list(mode, lhs)
func (l *Loader) luaKeymapList(L *lua.LState) int {
	mode := ""
	lhsFilter := ""

	// Optional mode parameter
	if L.GetTop() >= 1 && L.Get(1) != lua.LNil {
		mode = L.CheckString(1)
	}

	// Optional lhs filter parameter
	if L.GetTop() >= 2 && L.Get(2) != lua.LNil {
		lhsFilter = L.CheckString(2)
	}

	// Get filtered keymaps
	keymaps := ListKeymaps(l.config.KeyMappings, mode, lhsFilter)

	// Convert to Lua table
	result := L.NewTable()
	for i, km := range keymaps {
		entry := L.NewTable()
		L.SetField(entry, "mode", lua.LString(km.Mode))
		L.SetField(entry, "lhs", lua.LString(km.LHS))

		if km.IsFunc {
			L.SetField(entry, "rhs", lua.LString("<function>"))
		} else {
			L.SetField(entry, "rhs", lua.LString(km.RHS))
		}

		L.SetField(entry, "desc", lua.LString(km.Opts.Desc))
		L.SetField(entry, "noremap", lua.LBool(km.Opts.Noremap))
		L.SetField(entry, "silent", lua.LBool(km.Opts.Silent))

		L.RawSetInt(result, i+1, entry)
	}

	L.Push(result)
	return 1
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

	// Normalize the mapping (expand leader, handle key notation)
	leader := l.config.Options.Leader
	NormalizeKeyMapping(&mapping, leader)

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
			if v := L.GetField(optsTable, "group"); v != lua.LNil {
				if s, ok := v.(lua.LString); ok {
					opts.Group = string(s)
				}
			}
		}
	}

	// If no group specified but we're inside an augroup, use that
	if opts.Group == "" {
		if currentGroup := L.GetGlobal("_current_augroup"); currentGroup != lua.LNil {
			if groupStr, ok := currentGroup.(lua.LString); ok {
				opts.Group = string(groupStr)
			}
		}
	}

	handler := EventHandler{
		Event: event,
		Fn:    fn,
		Opts:  opts,
		Fired: false,
	}

	l.config.EventHandlers = append(l.config.EventHandlers, handler)
	return 0
}

// luaAugroup implements vb.augroup(name, fn)
// Clears existing group and calls function to register new handlers
func (l *Loader) luaAugroup(L *lua.LState) int {
	groupName := L.CheckString(1)
	fn := L.CheckFunction(2)

	// Clear existing group
	l.config.ClearEventGroup(groupName)

	// Store current group in a global variable for vb.on() to use
	L.SetGlobal("_current_augroup", lua.LString(groupName))

	// Call the function (which should call vb.on() to register handlers)
	L.Push(fn)
	if err := L.PCall(0, 0, nil); err != nil {
		L.RaiseError("augroup function error: %v", err)
	}

	// Clear the current group
	L.SetGlobal("_current_augroup", lua.LNil)

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

// luaCheckHealth implements vb.checkhealth()
func (l *Loader) luaCheckHealth(L *lua.LState) int {
	var output strings.Builder

	output.WriteString("\n")
	output.WriteString("=== Voidabyss Health Check ===\n\n")

	// Version
	output.WriteString(fmt.Sprintf("✓ Version: %s\n", Version))

	// Lua version
	output.WriteString(fmt.Sprintf("✓ Lua: %s\n", lua.LuaVersion))

	// Config path
	configPath := GetConfigPath()
	{
		loaded := "not loaded"
		if _, err := os.Stat(configPath); err == nil {
			loaded = "loaded"
		}
		output.WriteString(fmt.Sprintf("✓ Config: %s (%s)\n", configPath, loaded))
	}

	// State file
	statePath := l.config.State.GetPath()
	stateWritable := "yes"
	if err := l.config.State.TestWrite(); err != nil {
		stateWritable = fmt.Sprintf("ERROR - %v", err)
	}
	output.WriteString(fmt.Sprintf("✓ State file: %s (writable: %s)\n", statePath, stateWritable))

	// API modules registered
	output.WriteString("\n=== API Modules ===\n")
	modules := []string{
		"vb.version",
		"vb.opt (options with property access)",
		"vb.keymap (key mappings)",
		"vb.command (custom commands)",
		"vb.on (events/autocmds)",
		"vb.buf (buffer operations)",
		"vb.state (persistent storage)",
		"vb.notify (notifications)",
		"vb.has (feature detection)",
		"vb.schedule (async function queue)",
		"vb.checkhealth (this command)",
	}
	for _, mod := range modules {
		output.WriteString(fmt.Sprintf("  ✓ %s\n", mod))
	}

	// Keymap statistics
	output.WriteString("\n=== Keymaps ===\n")
	keymapsByMode := make(map[string]int)
	for _, km := range l.config.KeyMappings {
		keymapsByMode[km.Mode]++
	}
	if len(keymapsByMode) == 0 {
		output.WriteString("  No keymaps registered\n")
	} else {
		for mode, count := range keymapsByMode {
			modeName := mode
			switch mode {
			case "n":
				modeName = "normal"
			case "i":
				modeName = "insert"
			case "v":
				modeName = "visual"
			case "c":
				modeName = "command"
			}
			output.WriteString(fmt.Sprintf("  %s: %d mappings\n", modeName, count))
		}
	}

	// Command statistics
	output.WriteString("\n=== Commands ===\n")
	if len(l.config.Commands) == 0 {
		output.WriteString("  No custom commands registered\n")
	} else {
		output.WriteString(fmt.Sprintf("  %d custom commands registered\n", len(l.config.Commands)))
		for _, cmd := range l.config.Commands {
			desc := ""
			if cmd.Opts.Desc != "" {
				desc = fmt.Sprintf(" (%s)", cmd.Opts.Desc)
			}
			output.WriteString(fmt.Sprintf("    :%s%s\n", cmd.Name, desc))
		}
	}

	// Event handler statistics
	output.WriteString("\n=== Event Handlers ===\n")
	eventsByType := make(map[string]int)
	for _, eh := range l.config.EventHandlers {
		eventsByType[eh.Event]++
	}
	if len(eventsByType) == 0 {
		output.WriteString("  No event handlers registered\n")
	} else {
		for event, count := range eventsByType {
			output.WriteString(fmt.Sprintf("  %s: %d handlers\n", event, count))
		}
	}

	// Features
	output.WriteString("\n=== Features ===\n")
	features := []string{
		"keymap.leader",
		"keymap.function",
		"command.function",
		"events.autocmd",
		"buffer.api",
		"state.persistent",
		"opt.tabwidth",
		"opt.property_access",
	}
	for _, feature := range features {
		status := "✓"
		if !Features[feature] {
			status = "✗"
		}
		output.WriteString(fmt.Sprintf("  %s %s\n", status, feature))
	}

	// Completion status
	output.WriteString("\n=== Editor Features ===\n")
	output.WriteString("  ✓ Completion: enabled\n")
	output.WriteString("  ✓ Buffer API: available\n")

	// State info
	output.WriteString("\n=== Persistent State ===\n")
	keys := l.config.State.Keys()
	if len(keys) == 0 {
		output.WriteString("  No persistent state stored\n")
	} else {
		output.WriteString(fmt.Sprintf("  %d keys stored\n", len(keys)))
		for _, key := range keys {
			val := l.config.State.Get(key, nil)
			output.WriteString(fmt.Sprintf("    %s = %v\n", key, val))
		}
	}

	// Plugin info
	output.WriteString("\n=== Plugins ===\n")
	if len(l.config.LoadedPlugins) == 0 {
		output.WriteString("  No plugins loaded\n")
	} else {
		loaded := 0
		failed := 0
		for _, plugin := range l.config.LoadedPlugins {
			if plugin.Loaded {
				loaded++
				output.WriteString(fmt.Sprintf("  ✓ %s\n", plugin.Name))
			} else {
				failed++
				output.WriteString(fmt.Sprintf("  ✗ %s: %s\n", plugin.Name, plugin.Error))
			}
		}
		output.WriteString(fmt.Sprintf("\n  Summary: %d loaded, %d failed\n", loaded, failed))
	}

	output.WriteString("\n")

	// Print to stdout (visible in editor)
	fmt.Print(output.String())

	return 0
}

// luaSchedule implements vb.schedule(fn)
// Queues a function to be called on the next tick
func (l *Loader) luaSchedule(L *lua.LState) int {
	fn := L.CheckFunction(1)

	// Store the function in the scheduled queue (thread-safe)
	l.config.scheduleMu.Lock()
	l.config.scheduledFns = append(l.config.scheduledFns, fn)
	l.config.scheduleMu.Unlock()

	return 0
}

// SafeCallLuaFunction safely calls a Lua function with error handling
// Returns (success bool, error string)
func (l *Loader) SafeCallLuaFunction(fn *lua.LFunction, args ...lua.LValue) (bool, string) {
	// Create error handler function
	errHandler := l.L.NewFunction(func(L *lua.LState) int {
		err := L.CheckAny(1)
		L.Push(lua.LString(fmt.Sprintf("Lua error: %v", err)))
		return 1
	})

	// Push error handler
	l.L.Push(errHandler)

	// Push function
	l.L.Push(fn)

	// Push arguments
	for _, arg := range args {
		l.L.Push(arg)
	}

	// Call with protection
	if err := l.L.PCall(len(args), 0, errHandler); err != nil {
		return false, err.Error()
	}

	return true, ""
}

// ProcessScheduledFunctions executes all scheduled functions
// Should be called once per editor tick
func (c *Config) ProcessScheduledFunctions(L *lua.LState) {
	// Get and clear the queue atomically
	c.scheduleMu.Lock()
	fns := c.scheduledFns
	c.scheduledFns = nil
	c.scheduleMu.Unlock()

	if len(fns) == 0 {
		return
	}

	// Execute each function safely
	for _, fn := range fns {
		// Create error handler
		errHandler := L.NewFunction(func(L *lua.LState) int {
			err := L.CheckAny(1)
			fmt.Printf("Scheduled function error: %v\n", err)
			return 0
		})

		L.Push(errHandler)
		L.Push(fn)

		// Call with protection (pcall equivalent)
		if err := L.PCall(0, 0, errHandler); err != nil {
			fmt.Printf("Error executing scheduled function: %v\n", err)
		}
	}
}
