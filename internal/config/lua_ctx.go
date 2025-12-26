package config

import (
	lua "github.com/yuin/gopher-lua"
)

const luaCtxTypeName = "vb.ctx"

// LuaContext represents the context passed to Lua callbacks
type LuaContext struct {
	editor EditorContext
}

// RegisterCtxType registers the ctx userdata type with methods
func (l *Loader) RegisterCtxType() {
	mt := l.L.NewTypeMetatable(luaCtxTypeName)
	l.L.SetGlobal(luaCtxTypeName, mt)

	// Methods
	l.L.SetField(mt, "__index", l.L.SetFuncs(l.L.NewTable(), map[string]lua.LGFunction{
		// Completion methods
		"complete_next":   l.ctxCompleteNext,
		"complete_prev":   l.ctxCompletePrev,
		"complete_cancel": l.ctxCompleteCancel,

		// Editor methods
		"mode":       l.ctxMode,
		"set_mode":   l.ctxSetMode,
		"cursor":     l.ctxCursor,
		"set_cursor": l.ctxSetCursor,

		// Buffer methods (convenience wrappers)
		"get_text": l.ctxGetText,
		"set_text": l.ctxSetText,

		// Command execution
		"cmd":      l.ctxCmd,
		"feedkeys": l.ctxFeedkeys,

		// Info
		"file": l.ctxFile,
		"line": l.ctxLine,
		"col":  l.ctxCol,
	}))
}

// NewLuaContext creates a new Lua context userdata
func (l *Loader) NewLuaContext() *lua.LUserData {
	ctx := &LuaContext{
		editor: l.editorCtx,
	}
	ud := l.L.NewUserData()
	ud.Value = ctx
	l.L.SetMetatable(ud, l.L.GetTypeMetatable(luaCtxTypeName))
	return ud
}

// checkCtx validates and extracts LuaContext from userdata
func (l *Loader) checkCtx(L *lua.LState, n int) *LuaContext {
	ud := L.CheckUserData(n)
	if v, ok := ud.Value.(*LuaContext); ok {
		return v
	}
	L.ArgError(n, "ctx expected")
	return nil
}

// Completion methods
func (l *Loader) ctxCompleteNext(L *lua.LState) int {
	ctx := l.checkCtx(L, 1)
	if completer, ok := ctx.editor.(interface{ CompleteNext() }); ok {
		completer.CompleteNext()
	}
	return 0
}

func (l *Loader) ctxCompletePrev(L *lua.LState) int {
	ctx := l.checkCtx(L, 1)
	if completer, ok := ctx.editor.(interface{ CompletePrev() }); ok {
		completer.CompletePrev()
	}
	return 0
}

func (l *Loader) ctxCompleteCancel(L *lua.LState) int {
	ctx := l.checkCtx(L, 1)
	if completer, ok := ctx.editor.(interface{ CompleteCancel() }); ok {
		completer.CompleteCancel()
	}
	return 0
}

// Editor methods
func (l *Loader) ctxMode(L *lua.LState) int {
	ctx := l.checkCtx(L, 1)
	if ctx.editor != nil {
		mode := ctx.editor.Mode()
		L.Push(lua.LString(mode))
		return 1
	}
	L.Push(lua.LString("normal"))
	return 1
}

func (l *Loader) ctxSetMode(L *lua.LState) int {
	ctx := l.checkCtx(L, 1)
	mode := L.CheckString(2)
	if ctx.editor != nil {
		ctx.editor.SetMode(mode)
	}
	return 0
}

func (l *Loader) ctxCursor(L *lua.LState) int {
	ctx := l.checkCtx(L, 1)
	if ctx.editor != nil {
		row, col := ctx.editor.Cursor()
		L.Push(lua.LNumber(row))
		L.Push(lua.LNumber(col))
		return 2
	}
	L.Push(lua.LNumber(0))
	L.Push(lua.LNumber(0))
	return 2
}

func (l *Loader) ctxSetCursor(L *lua.LState) int {
	ctx := l.checkCtx(L, 1)
	row := L.CheckInt(2)
	col := L.CheckInt(3)
	if ctx.editor != nil {
		ctx.editor.SetCursor(row, col)
	}
	return 0
}

// Buffer convenience methods
func (l *Loader) ctxGetText(L *lua.LState) int {
	ctx := l.checkCtx(L, 1)
	if ctx.editor != nil {
		text := ctx.editor.GetText()
		L.Push(lua.LString(text))
		return 1
	}
	L.Push(lua.LString(""))
	return 1
}

func (l *Loader) ctxSetText(L *lua.LState) int {
	ctx := l.checkCtx(L, 1)
	text := L.CheckString(2)
	if ctx.editor != nil {
		ctx.editor.SetText(text)
	}
	return 0
}

// Command execution methods
func (l *Loader) ctxCmd(L *lua.LState) int {
	ctx := l.checkCtx(L, 1)
	cmd := L.CheckString(2)
	if execer, ok := ctx.editor.(interface{ ExecCommand(string) bool }); ok {
		quit := execer.ExecCommand(cmd)
		L.Push(lua.LBool(quit))
		return 1
	}
	L.Push(lua.LBool(false))
	return 1
}

func (l *Loader) ctxFeedkeys(L *lua.LState) int {
	ctx := l.checkCtx(L, 1)
	keys := L.CheckString(2)
	if feeder, ok := ctx.editor.(interface{ Feedkeys(string) }); ok {
		feeder.Feedkeys(keys)
	}
	return 0
}

// Info methods
func (l *Loader) ctxFile(L *lua.LState) int {
	ctx := l.checkCtx(L, 1)
	if ctx.editor != nil {
		name := ctx.editor.GetName()
		L.Push(lua.LString(name))
		return 1
	}
	L.Push(lua.LString(""))
	return 1
}

func (l *Loader) ctxLine(L *lua.LState) int {
	ctx := l.checkCtx(L, 1)
	if ctx.editor != nil {
		row, _ := ctx.editor.Cursor()
		L.Push(lua.LNumber(row))
		return 1
	}
	L.Push(lua.LNumber(0))
	return 1
}

func (l *Loader) ctxCol(L *lua.LState) int {
	ctx := l.checkCtx(L, 1)
	if ctx.editor != nil {
		_, col := ctx.editor.Cursor()
		L.Push(lua.LNumber(col))
		return 1
	}
	L.Push(lua.LNumber(0))
	return 1
}
