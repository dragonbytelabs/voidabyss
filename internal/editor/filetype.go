package editor

import (
	"path/filepath"
	"strings"
)

// Filetype represents a detected file type
type Filetype struct {
	Name       string // e.g., "go", "python", "javascript"
	Extensions []string
	Comment    string // comment prefix (e.g., "//" for Go, "#" for Python)
}

var filetypes = []Filetype{
	{
		Name:       "go",
		Extensions: []string{".go"},
		Comment:    "//",
	},
	{
		Name:       "python",
		Extensions: []string{".py", ".pyw"},
		Comment:    "#",
	},
	{
		Name:       "javascript",
		Extensions: []string{".js", ".jsx", ".mjs"},
		Comment:    "//",
	},
	{
		Name:       "typescript",
		Extensions: []string{".ts", ".tsx"},
		Comment:    "//",
	},
	{
		Name:       "c",
		Extensions: []string{".c", ".h"},
		Comment:    "//",
	},
	{
		Name:       "cpp",
		Extensions: []string{".cpp", ".cc", ".cxx", ".hpp", ".hxx", ".h++"},
		Comment:    "//",
	},
	{
		Name:       "rust",
		Extensions: []string{".rs"},
		Comment:    "//",
	},
	{
		Name:       "ruby",
		Extensions: []string{".rb"},
		Comment:    "#",
	},
	{
		Name:       "php",
		Extensions: []string{".php"},
		Comment:    "//",
	},
	{
		Name:       "java",
		Extensions: []string{".java"},
		Comment:    "//",
	},
	{
		Name:       "lua",
		Extensions: []string{".lua"},
		Comment:    "--",
	},
	{
		Name:       "shell",
		Extensions: []string{".sh", ".bash", ".zsh"},
		Comment:    "#",
	},
	{
		Name:       "yaml",
		Extensions: []string{".yml", ".yaml"},
		Comment:    "#",
	},
	{
		Name:       "json",
		Extensions: []string{".json"},
		Comment:    "",
	},
	{
		Name:       "markdown",
		Extensions: []string{".md", ".markdown"},
		Comment:    "",
	},
	{
		Name:       "html",
		Extensions: []string{".html", ".htm"},
		Comment:    "<!--",
	},
	{
		Name:       "css",
		Extensions: []string{".css"},
		Comment:    "/*",
	},
	{
		Name:       "xml",
		Extensions: []string{".xml"},
		Comment:    "<!--",
	},
	{
		Name:       "sql",
		Extensions: []string{".sql"},
		Comment:    "--",
	},
	{
		Name:       "vim",
		Extensions: []string{".vim"},
		Comment:    "\"",
	},
	{
		Name:       "toml",
		Extensions: []string{".toml"},
		Comment:    "#",
	},
}

// detectFiletype detects the filetype from filename
func detectFiletype(filename string) *Filetype {
	if filename == "" {
		return nil
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		// Check for special files without extensions
		base := strings.ToLower(filepath.Base(filename))
		switch base {
		case "makefile":
			return &Filetype{Name: "make", Comment: "#"}
		case "dockerfile":
			return &Filetype{Name: "dockerfile", Comment: "#"}
		case "rakefile":
			return &Filetype{Name: "ruby", Comment: "#"}
		case "gemfile":
			return &Filetype{Name: "ruby", Comment: "#"}
		}
		return nil
	}

	for _, ft := range filetypes {
		for _, ftExt := range ft.Extensions {
			if ext == ftExt {
				return &ft
			}
		}
	}

	return nil
}

// getFiletype returns the current buffer's filetype
func (e *Editor) getFiletype() *Filetype {
	return detectFiletype(e.filename)
}

// setFileTypeOptions applies filetype-specific settings
func (e *Editor) setFiletypeOptions() {
	ft := e.getFiletype()
	if ft == nil {
		return
	}

	// Apply language-specific indent settings
	switch ft.Name {
	case "go":
		// Go uses tabs
		e.config.TabWidth = 4
	case "python", "yaml", "ruby":
		// These typically use 4 spaces
		e.indentWidth = 4
	case "javascript", "typescript", "json", "html", "css":
		// These typically use 2 spaces
		e.indentWidth = 2
	}

	// Initialize tree-sitter parser for supported languages
	e.initTreeSitterParser()
}

// initTreeSitterParser initializes tree-sitter parser for the current buffer
func (e *Editor) initTreeSitterParser() {
	ft := e.getFiletype()
	if ft == nil {
		return
	}

	// Get current buffer view
	bv := e.buf()
	if bv == nil {
		return
	}

	// Close existing parser if any
	if bv.parser != nil {
		bv.parser.Close()
		bv.parser = nil
	}

	// Create new parser for supported languages
	parser, err := NewTreeSitterParser(ft.Name)
	if err != nil || parser == nil {
		// Language not supported or error creating parser
		return
	}

	// Parse current buffer content
	content := e.buffer.String()
	if err := parser.Parse(content); err != nil {
		parser.Close()
		return
	}

	bv.parser = parser
}

// reparseBuffer re-parses the buffer after modifications
func (e *Editor) reparseBuffer() {
	bv := e.buf()
	if bv == nil || bv.parser == nil {
		return
	}

	content := e.buffer.String()
	bv.parser.Parse(content)
}

// getCommentPrefix returns the comment prefix for the current filetype
func (e *Editor) getCommentPrefix() string {
	ft := e.getFiletype()
	if ft == nil {
		return "//"
	}
	if ft.Comment == "" {
		return "//"
	}
	return ft.Comment
}
