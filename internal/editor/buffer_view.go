package editor

import (
	"github.com/dragonbytelabs/voidabyss/core/buffer"
)

// BufferView represents a single file/buffer being edited
type BufferView struct {
	buffer   *buffer.Buffer
	filename string
	dirty    bool

	// cursor position in (line, col)
	cx, cy int

	// viewport offsets
	rowOffset int
	colOffset int

	// cursor column preference for vertical motion
	wantX int

	// marks specific to this buffer
	marks map[rune]Mark

	// jump list specific to this buffer
	jumpList      []JumpListEntry
	jumpListIndex int

	// tree-sitter parser for syntax highlighting
	parser *TreeSitterParser
}

// NewBufferView creates a new buffer view from content and filename
func NewBufferView(content, filename string) *BufferView {
	return &BufferView{
		buffer:        buffer.NewFromString(content),
		filename:      filename,
		dirty:         false,
		marks:         make(map[rune]Mark),
		jumpList:      make([]JumpListEntry, 0, 100),
		jumpListIndex: -1,
	}
}
