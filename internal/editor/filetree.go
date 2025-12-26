package editor

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileTreeNode represents a file or directory in the tree
type FileTreeNode struct {
	path     string
	name     string
	isDir    bool
	children []*FileTreeNode
	expanded bool
	parent   *FileTreeNode
}

// FileTree manages the file explorer
type FileTree struct {
	root        *FileTreeNode
	flat        []*FileTreeNode // flattened visible nodes
	cursor      int             // current selection index in flat list
	rootPath    string
	ignoreRules []string
}

var defaultIgnoreRules = []string{
	".git",
	"node_modules",
	"dist",
	"build",
	"target",
	".next",
	".cache",
	"__pycache__",
	".pytest_cache",
	"vendor",
	".DS_Store",
}

// NewFileTree creates a new file tree rooted at the given path
func NewFileTree(rootPath string) (*FileTree, error) {
	abs, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}

	// If it's a file, use its directory
	if !info.IsDir() {
		abs = filepath.Dir(abs)
	}

	ft := &FileTree{
		rootPath:    abs,
		ignoreRules: defaultIgnoreRules,
		cursor:      0,
	}

	// Build the tree
	root, err := ft.buildNode(abs, nil)
	if err != nil {
		return nil, err
	}
	root.expanded = true
	ft.root = root

	// Flatten the tree for display
	ft.rebuildFlat()

	return ft, nil
}

// buildNode recursively builds a tree node
func (ft *FileTree) buildNode(path string, parent *FileTreeNode) (*FileTreeNode, error) {
	name := filepath.Base(path)
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	node := &FileTreeNode{
		path:   path,
		name:   name,
		isDir:  info.IsDir(),
		parent: parent,
	}

	if node.isDir {
		entries, err := os.ReadDir(path)
		if err != nil {
			return node, nil // Return node even if we can't read dir
		}

		for _, entry := range entries {
			// Skip ignored files/dirs
			if ft.shouldIgnore(entry.Name()) {
				continue
			}

			childPath := filepath.Join(path, entry.Name())
			child, err := ft.buildNode(childPath, node)
			if err != nil {
				continue // Skip problematic entries
			}
			node.children = append(node.children, child)
		}

		// Sort: directories first, then files, both alphabetically
		sort.Slice(node.children, func(i, j int) bool {
			if node.children[i].isDir != node.children[j].isDir {
				return node.children[i].isDir
			}
			return strings.ToLower(node.children[i].name) < strings.ToLower(node.children[j].name)
		})
	}

	return node, nil
}

// shouldIgnore checks if a filename should be ignored
func (ft *FileTree) shouldIgnore(name string) bool {
	for _, rule := range ft.ignoreRules {
		if name == rule {
			return true
		}
		// Basic glob matching for patterns like *.log
		if strings.HasPrefix(rule, "*.") {
			ext := rule[1:]
			if strings.HasSuffix(name, ext) {
				return true
			}
		}
	}
	return false
}

// rebuildFlat creates a flat list of visible nodes
func (ft *FileTree) rebuildFlat() {
	ft.flat = make([]*FileTreeNode, 0)
	if ft.root != nil {
		ft.flattenNode(ft.root, 0)
	}
}

// flattenNode recursively flattens visible nodes
func (ft *FileTree) flattenNode(node *FileTreeNode, depth int) {
	ft.flat = append(ft.flat, node)
	if node.isDir && node.expanded {
		for _, child := range node.children {
			ft.flattenNode(child, depth+1)
		}
	}
}

// ensureCursorValid ensures cursor is within valid range
func (ft *FileTree) ensureCursorValid() {
	if ft.cursor >= len(ft.flat) {
		ft.cursor = len(ft.flat) - 1
	}
	if ft.cursor < 0 {
		ft.cursor = 0
	}
}

// getCurrentNode returns the currently selected node
func (ft *FileTree) getCurrentNode() *FileTreeNode {
	if ft.cursor >= 0 && ft.cursor < len(ft.flat) {
		return ft.flat[ft.cursor]
	}
	return nil
}

// moveUp moves cursor up
func (ft *FileTree) moveUp() {
	if ft.cursor > 0 {
		ft.cursor--
	}
}

// moveDown moves cursor down
func (ft *FileTree) moveDown() {
	if ft.cursor < len(ft.flat)-1 {
		ft.cursor++
	}
}

// getDepth returns the depth of a node in the tree
func (ft *FileTree) getDepth(node *FileTreeNode) int {
	depth := 0
	current := node.parent
	for current != nil {
		depth++
		current = current.parent
	}
	return depth
}

// getDisplayLines returns the lines to display in the tree panel
func (ft *FileTree) getDisplayLines() []string {
	lines := make([]string, len(ft.flat))
	for i, node := range ft.flat {
		depth := ft.getDepth(node)
		indent := strings.Repeat("  ", depth)

		prefix := " "
		if node.isDir {
			if node.expanded {
				prefix = "▼"
			} else {
				prefix = "▶"
			}
		}

		cursor := " "
		if i == ft.cursor {
			cursor = ">"
		}

		lines[i] = cursor + indent + prefix + " " + node.name
	}
	return lines
}
