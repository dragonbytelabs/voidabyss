package config

import (
	"encoding/json"

	lua "github.com/yuin/gopher-lua"
)

// LSPClientRegistry stores LSP clients created by Lua plugins
var lspClientRegistry = make(map[string]interface{})

// RegisterLSPAPI adds LSP-related functions to the Lua state
func RegisterLSPAPI(L *lua.LState) {
	// Create lsp table
	lspTable := L.NewTable()

	// lsp.start_client(config)
	L.SetField(lspTable, "start_client", L.NewFunction(luaStartLSPClient))

	// lsp.stop_client(client_id)
	L.SetField(lspTable, "stop_client", L.NewFunction(luaStopLSPClient))

	// lsp.goto_definition(filepath, line, col, callback)
	L.SetField(lspTable, "goto_definition", L.NewFunction(luaGotoDefinition))

	// lsp.hover(filepath, line, col, callback)
	L.SetField(lspTable, "hover", L.NewFunction(luaHover))

	// lsp.did_open(client_id, filepath, content)
	L.SetField(lspTable, "did_open", L.NewFunction(luaDidOpen))

	// lsp.did_change(client_id, filepath, content)
	L.SetField(lspTable, "did_change", L.NewFunction(luaDidChange))

	// lsp.did_save(client_id, filepath)
	L.SetField(lspTable, "did_save", L.NewFunction(luaDidSave))

	// lsp.did_close(client_id, filepath)
	L.SetField(lspTable, "did_close", L.NewFunction(luaDidClose))

	L.SetGlobal("lsp", lspTable)
}

// luaStartLSPClient starts an LSP client
// Usage: client_id = lsp.start_client({cmd = "gopls", args = {}, root_dir = "/path"})
func luaStartLSPClient(L *lua.LState) int {
	config := L.CheckTable(1)

	cmd := config.RawGetString("cmd").String()
	rootDir := config.RawGetString("root_dir").String()

	// Extract args
	var args []string
	if argsVal := config.RawGetString("args"); argsVal != lua.LNil {
		if argsTable, ok := argsVal.(*lua.LTable); ok {
			argsTable.ForEach(func(_ lua.LValue, val lua.LValue) {
				args = append(args, val.String())
			})
		}
	}

	// This will be implemented to actually start the LSP client
	// For now, just return a placeholder
	clientID := cmd + ":" + rootDir
	lspClientRegistry[clientID] = map[string]interface{}{
		"cmd":      cmd,
		"args":     args,
		"root_dir": rootDir,
	}

	L.Push(lua.LString(clientID))
	return 1
}

// luaStopLSPClient stops an LSP client
func luaStopLSPClient(L *lua.LState) int {
	clientID := L.CheckString(1)
	delete(lspClientRegistry, clientID)
	return 0
}

// luaGotoDefinition requests definition location from LSP
// Usage: lsp.goto_definition(client_id, filepath, line, col, function(locations) ... end)
func luaGotoDefinition(L *lua.LState) int {
	clientID := L.CheckString(1)
	filepath := L.CheckString(2)
	line := L.CheckInt(3)
	col := L.CheckInt(4)
	callback := L.CheckFunction(5)

	// This will be implemented to actually call LSP
	// For now, just call the callback with empty result
	go func() {
		locationsTable := L.NewTable()
		L.Push(callback)
		L.Push(locationsTable)
		if err := L.PCall(1, 0, nil); err != nil {
			L.RaiseError("goto_definition callback error: %v", err)
		}
	}()

	// Silence unused variable warnings
	_, _, _, _ = clientID, filepath, line, col

	return 0
}

// luaHover requests hover information from LSP
func luaHover(L *lua.LState) int {
	clientID := L.CheckString(1)
	filepath := L.CheckString(2)
	line := L.CheckInt(3)
	col := L.CheckInt(4)
	callback := L.CheckFunction(5)

	// Will be implemented later
	_, _, _, _, _ = clientID, filepath, line, col, callback
	return 0
}

// luaDidOpen notifies LSP of opened document
func luaDidOpen(L *lua.LState) int {
	clientID := L.CheckString(1)
	filepath := L.CheckString(2)
	content := L.CheckString(3)

	// Will be implemented later
	_, _, _ = clientID, filepath, content
	return 0
}

// luaDidChange notifies LSP of document changes
func luaDidChange(L *lua.LState) int {
	clientID := L.CheckString(1)
	filepath := L.CheckString(2)
	content := L.CheckString(3)

	// Will be implemented later
	_, _, _ = clientID, filepath, content
	return 0
}

// luaDidSave notifies LSP of document save
func luaDidSave(L *lua.LState) int {
	clientID := L.CheckString(1)
	filepath := L.CheckString(2)

	// Will be implemented later
	_, _ = clientID, filepath
	return 0
}

// luaDidClose notifies LSP of closed document
func luaDidClose(L *lua.LState) int {
	clientID := L.CheckString(1)
	filepath := L.CheckString(2)

	// Will be implemented later
	_, _ = clientID, filepath
	return 0
}

// Helper to convert Lua table to JSON (for future use)
func luaTableToJSON(L *lua.LState, table *lua.LTable) (string, error) {
	data := luaTableToMap(table)
	bytes, err := json.Marshal(data)
	return string(bytes), err
}

func luaTableToMap(table *lua.LTable) map[string]interface{} {
	result := make(map[string]interface{})
	table.ForEach(func(key, val lua.LValue) {
		keyStr := key.String()
		switch v := val.(type) {
		case *lua.LTable:
			result[keyStr] = luaTableToMap(v)
		case lua.LString:
			result[keyStr] = string(v)
		case lua.LNumber:
			result[keyStr] = float64(v)
		case lua.LBool:
			result[keyStr] = bool(v)
		}
	})
	return result
}
