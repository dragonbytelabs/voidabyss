package editor

import (
	"sync"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

// TreeSitterParser manages tree-sitter parsing for a buffer
type TreeSitterParser struct {
	parser   *sitter.Parser
	tree     *sitter.Tree
	language *sitter.Language
	mu       sync.RWMutex
}

// NewTreeSitterParser creates a new parser for the given language
func NewTreeSitterParser(langName string) (*TreeSitterParser, error) {
	parser := sitter.NewParser()

	var language *sitter.Language
	switch langName {
	case "go":
		language = sitter.NewLanguage(tree_sitter_go.Language())
	case "python":
		language = sitter.NewLanguage(tree_sitter_python.Language())
	case "javascript", "typescript", "jsx", "tsx":
		language = sitter.NewLanguage(tree_sitter_javascript.Language())
	default:
		// Return nil parser for unsupported languages
		return nil, nil
	}

	if err := parser.SetLanguage(language); err != nil {
		return nil, err
	}

	return &TreeSitterParser{
		parser:   parser,
		language: language,
	}, nil
}

// Parse parses the given source code
func (p *TreeSitterParser) Parse(source string) error {
	if p == nil || p.parser == nil {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	sourceBytes := []byte(source)

	tree := p.parser.Parse(sourceBytes, p.tree)

	p.tree = tree
	return nil
}

// GetTree returns the current parse tree
func (p *TreeSitterParser) GetTree() *sitter.Tree {
	if p == nil {
		return nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.tree
}

// GetHighlights returns syntax highlighting information for the visible range
func (p *TreeSitterParser) GetHighlights(startByte, endByte int) []Highlight {
	if p == nil || p.tree == nil {
		return nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	root := p.tree.RootNode()
	if root == nil {
		return nil
	}

	var highlights []Highlight
	p.collectHighlights(root, startByte, endByte, &highlights)
	return highlights
}

// Highlight represents a syntax highlight span
type Highlight struct {
	StartByte int
	EndByte   int
	Type      HighlightType
}

// HighlightType represents different syntax element types
type HighlightType int

const (
	HighlightNone HighlightType = iota
	HighlightKeyword
	HighlightFunction
	HighlightType_
	HighlightString
	HighlightNumber
	HighlightComment
	HighlightVariable
	HighlightOperator
	HighlightConstant
	HighlightProperty
)

// collectHighlights recursively collects highlights from the tree
func (p *TreeSitterParser) collectHighlights(node *sitter.Node, startByte, endByte int, highlights *[]Highlight) {
	if node == nil {
		return
	}

	nodeStart := int(node.StartByte())
	nodeEnd := int(node.EndByte())

	// Skip nodes outside the range
	if nodeEnd < startByte || nodeStart > endByte {
		return
	}

	// Get highlight type for this node
	nodeType := node.GrammarName()
	highlightType := p.getHighlightType(nodeType)

	if highlightType != HighlightNone {
		*highlights = append(*highlights, Highlight{
			StartByte: nodeStart,
			EndByte:   nodeEnd,
			Type:      highlightType,
		})
	}

	// Process children
	childCount := node.ChildCount()
	for i := uint(0); i < childCount; i++ {
		child := node.Child(i)
		p.collectHighlights(child, startByte, endByte, highlights)
	}
}

// getHighlightType maps tree-sitter node types to highlight types
func (p *TreeSitterParser) getHighlightType(nodeType string) HighlightType {
	// Keywords
	keywords := map[string]bool{
		"func": true, "package": true, "import": true, "var": true,
		"const": true, "type": true, "struct": true, "interface": true,
		"if": true, "else": true, "for": true, "range": true,
		"return": true, "break": true, "continue": true, "switch": true,
		"case": true, "default": true, "go": true, "defer": true,
		"select": true, "chan": true, "map": true,
		// Python
		"def": true, "class": true, "lambda": true, "pass": true,
		"with": true, "as": true, "yield": true, "async": true, "await": true,
		"raise": true, "raise_from": true, "try": true, "except": true, "finally": true,
		"while": true, "in": true, "is": true, "not": true, "and": true, "or": true,
		// JavaScript
		"function": true, "let": true, "const_decl": true, "new": true,
		"this": true, "super": true, "extends": true, "implements": true,
		"typeof": true, "instanceof": true, "delete": true,
		"export": true, "import_statement": true, "from_import": true, "default_export": true,
	}

	if keywords[nodeType] {
		return HighlightKeyword
	}

	// Node type patterns
	switch nodeType {
	case "function_declaration", "function_definition", "method_declaration":
		return HighlightFunction
	case "type_identifier", "type_spec", "primitive_type":
		return HighlightType_
	case "string_literal", "interpreted_string_literal", "raw_string_literal", "string":
		return HighlightString
	case "int_literal", "float_literal", "number":
		return HighlightNumber
	case "comment", "line_comment", "block_comment":
		return HighlightComment
	case "identifier":
		return HighlightVariable
	case "const_spec", "const_declaration":
		return HighlightConstant
	case "field_identifier", "property_identifier":
		return HighlightProperty
	case "+", "-", "*", "/", "=", "==", "!=", "<", ">", "<=", ">=":
		return HighlightOperator
	}

	return HighlightNone
}

// Close releases resources
func (p *TreeSitterParser) Close() {
	if p == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.tree != nil {
		p.tree.Close()
		p.tree = nil
	}

	if p.parser != nil {
		p.parser.Close()
		p.parser = nil
	}
}
