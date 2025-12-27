package lsp

import (
	"fmt"
	"path/filepath"
)

// DocumentSync handles document synchronization with the language server
type DocumentSync struct {
	client  *Client
	version int
	uri     string
}

// NewDocumentSync creates a new document sync handler
func NewDocumentSync(client *Client, filepath string) *DocumentSync {
	uri := "file://" + filepath
	return &DocumentSync{
		client:  client,
		uri:     uri,
		version: 0,
	}
}

// DidOpen notifies the server that a document was opened
func (ds *DocumentSync) DidOpen(languageID, text string) error {
	ds.version = 1

	params := DidOpenTextDocumentParams{
		TextDocument: TextDocumentItem{
			URI:        ds.uri,
			LanguageID: languageID,
			Version:    ds.version,
			Text:       text,
		},
	}

	return ds.client.Notify("textDocument/didOpen", params)
}

// DidChange notifies the server that a document was changed
func (ds *DocumentSync) DidChange(text string) error {
	ds.version++

	// Full document sync (sending entire content)
	params := DidChangeTextDocumentParams{
		TextDocument: VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: TextDocumentIdentifier{
				URI: ds.uri,
			},
			Version: ds.version,
		},
		ContentChanges: []TextDocumentContentChangeEvent{
			{
				Text: text,
			},
		},
	}

	return ds.client.Notify("textDocument/didChange", params)
}

// DidSave notifies the server that a document was saved
func (ds *DocumentSync) DidSave(text string) error {
	params := DidSaveTextDocumentParams{
		TextDocument: TextDocumentIdentifier{
			URI: ds.uri,
		},
		Text: &text,
	}

	return ds.client.Notify("textDocument/didSave", params)
}

// DidClose notifies the server that a document was closed
func (ds *DocumentSync) DidClose() error {
	params := DidCloseTextDocumentParams{
		TextDocument: TextDocumentIdentifier{
			URI: ds.uri,
		},
	}

	return ds.client.Notify("textDocument/didClose", params)
}

// Definition requests the definition location for a position in the document
func (ds *DocumentSync) Definition(line, character int) ([]Location, error) {
	params := TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{
			URI: ds.uri,
		},
		Position: Position{
			Line:      line,
			Character: character,
		},
	}

	var result []Location
	if err := ds.client.Call("textDocument/definition", params, &result); err != nil {
		// Try single location format
		var singleResult Location
		if err := ds.client.Call("textDocument/definition", params, &singleResult); err != nil {
			return nil, fmt.Errorf("definition request: %w", err)
		}
		result = []Location{singleResult}
	}

	return result, nil
}

// GetLanguageID returns the language ID for a file extension
func GetLanguageID(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js", ".jsx":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".rs":
		return "rust"
	case ".c":
		return "c"
	case ".cpp", ".cc", ".cxx":
		return "cpp"
	case ".java":
		return "java"
	default:
		return ""
	}
}
