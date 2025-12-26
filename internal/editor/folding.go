package editor

import (
	"sort"

	sitter "github.com/tree-sitter/go-tree-sitter"
)

// FoldRange represents a foldable region
type FoldRange struct {
	startLine int  // 0-based line number
	endLine   int  // 0-based line number (inclusive)
	folded    bool // whether this range is currently folded
}

// GetFoldableRanges returns all foldable regions in the buffer using tree-sitter
func (e *Editor) GetFoldableRanges() []FoldRange {
	if e.parser == nil || e.parser.GetTree() == nil {
		return nil
	}

	tree := e.parser.GetTree()
	root := tree.RootNode()
	if root == nil {
		return nil
	}

	ranges := []FoldRange{}
	e.collectFoldableNodes(root, &ranges)

	// Sort by start line
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].startLine < ranges[j].startLine
	})

	return ranges
}

// collectFoldableNodes recursively collects foldable nodes from the AST
func (e *Editor) collectFoldableNodes(node *sitter.Node, ranges *[]FoldRange) {
	if node == nil {
		return
	}

	nodeType := node.GrammarName()
	startLine := int(node.StartPosition().Row)
	endLine := int(node.EndPosition().Row)

	// Only create folds for multi-line constructs
	if endLine > startLine {
		// Check if this is a foldable node type
		if e.isFoldableNodeType(nodeType) {
			*ranges = append(*ranges, FoldRange{
				startLine: startLine,
				endLine:   endLine,
				folded:    false,
			})
		}
	}

	// Recurse to children
	childCount := node.ChildCount()
	for i := uint(0); i < childCount; i++ {
		child := node.Child(i)
		if child != nil {
			e.collectFoldableNodes(child, ranges)
		}
	}
}

// isFoldableNodeType determines if a node type should be foldable
func (e *Editor) isFoldableNodeType(nodeType string) bool {
	// Common foldable node types across languages
	foldableTypes := map[string]bool{
		// Go
		"function_declaration": true,
		"method_declaration":   true,
		"type_declaration":     true,
		"block":                true,
		"struct_type":          true,
		"interface_type":       true,

		// Python
		"function_definition": true,
		"class_definition":    true,
		"while_statement":     true,
		"with_statement":      true,
		"try_statement":       true,

		// Common across languages (Go, Python, JS)
		"if_statement":  true,
		"for_statement": true,

		// JavaScript/TypeScript
		"function":              true,
		"arrow_function":        true,
		"class_declaration":     true,
		"method_definition":     true,
		"object":                true,
		"array":                 true,
		"statement_block":       true,
		"switch_statement":      true,
		"lexical_declaration":   true,
		"class":                 true,
		"export_statement":      true,
		"interface_declaration": true,
	}

	return foldableTypes[nodeType]
}

// UpdateFoldStates updates fold states from the existing foldRanges
func (e *Editor) UpdateFoldStates() {
	if e.foldRanges == nil {
		e.foldRanges = make(map[int]*FoldRange)
	}

	// Get fresh foldable ranges
	newRanges := e.GetFoldableRanges()

	// Preserve existing fold states
	newFoldMap := make(map[int]*FoldRange)
	for i := range newRanges {
		startLine := newRanges[i].startLine
		// Check if we had this fold before
		if oldFold, exists := e.foldRanges[startLine]; exists {
			newRanges[i].folded = oldFold.folded
		}
		newFoldMap[startLine] = &newRanges[i]
	}

	e.foldRanges = newFoldMap
}

// ToggleFold toggles the fold at the current cursor line
func (e *Editor) ToggleFold() {
	if e.foldRanges == nil {
		e.UpdateFoldStates()
	}

	// Find fold at or containing current line
	fold := e.findFoldAtLine(e.cy)
	if fold != nil {
		fold.folded = !fold.folded
		if fold.folded {
			e.statusMsg = "folded"
		} else {
			e.statusMsg = "unfolded"
		}
	} else {
		e.statusMsg = "no fold at cursor"
	}
}

// findFoldAtLine finds a fold that contains or starts at the given line
func (e *Editor) findFoldAtLine(line int) *FoldRange {
	// Look for a fold starting at this line first
	if fold, exists := e.foldRanges[line]; exists {
		return fold
	}

	// Search for any fold that contains this line
	// Find the innermost (smallest) fold containing the cursor
	var bestFold *FoldRange
	bestSize := 999999

	for _, fold := range e.foldRanges {
		if line >= fold.startLine && line <= fold.endLine {
			size := fold.endLine - fold.startLine
			if size < bestSize {
				bestSize = size
				bestFold = fold
			}
		}
	}

	return bestFold
}

// FoldAll folds all foldable regions
func (e *Editor) FoldAll() {
	if e.foldRanges == nil {
		e.UpdateFoldStates()
	}

	count := 0
	for _, fold := range e.foldRanges {
		if !fold.folded {
			fold.folded = true
			count++
		}
	}

	e.statusMsg = "folded all"
}

// UnfoldAll unfolds all foldable regions
func (e *Editor) UnfoldAll() {
	if e.foldRanges == nil {
		return
	}

	count := 0
	for _, fold := range e.foldRanges {
		if fold.folded {
			fold.folded = false
			count++
		}
	}

	e.statusMsg = "unfolded all"
}

// isLineFolded checks if a line is currently folded (hidden)
func (e *Editor) isLineFolded(line int) bool {
	if e.foldRanges == nil {
		return false
	}

	// Check if this line is inside a folded range
	for _, fold := range e.foldRanges {
		if fold.folded && line > fold.startLine && line <= fold.endLine {
			return true
		}
	}

	return false
}

// getVisibleLine converts a visual line number to actual line number
// accounting for folds
func (e *Editor) getVisibleLine(visualLine int) int {
	if e.foldRanges == nil || len(e.foldRanges) == 0 {
		return visualLine
	}

	actualLine := 0
	visibleCount := 0

	totalLines := e.lineCount()
	for actualLine < totalLines {
		if !e.isLineFolded(actualLine) {
			if visibleCount == visualLine {
				return actualLine
			}
			visibleCount++
		}
		actualLine++
	}

	return actualLine
}

// getVisibleLineCount returns the number of visible lines (excluding folded)
func (e *Editor) getVisibleLineCount() int {
	if e.foldRanges == nil || len(e.foldRanges) == 0 {
		return e.lineCount()
	}

	count := 0
	totalLines := e.lineCount()
	for line := 0; line < totalLines; line++ {
		if !e.isLineFolded(line) {
			count++
		}
	}

	return count
}

// getFoldIndicator returns a string indicating fold state for a line
func (e *Editor) getFoldIndicator(line int) string {
	if e.foldRanges == nil {
		return " "
	}

	if fold, exists := e.foldRanges[line]; exists {
		if fold.folded {
			return "▶" // Folded indicator
		}
		return "▼" // Foldable but not folded
	}

	return " "
}
