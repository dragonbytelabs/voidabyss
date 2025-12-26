package config

// EditorContext defines the interface that Lua callbacks can use to interact with the editor
type EditorContext interface {
	// Buffer operations
	GetText() string
	SetText(text string)
	GetName() string
	SetName(name string)

	// Cursor operations
	Cursor() (row, col int)
	SetCursor(row, col int)

	// Line operations
	Line(y int) string
	SetLine(y int, text string)

	// Editing operations
	Insert(pos int, text string)
	Delete(start, end int)
	Len() int

	// Mode operations (for future ctx)
	Mode() string
	SetMode(mode string)
}

// SetEditorContext sets the editor context for buffer operations
func (l *Loader) SetEditorContext(ctx EditorContext) {
	l.editorCtx = ctx
}

// Add field to Loader struct (will be added to loader.go)
// editorCtx EditorContext
