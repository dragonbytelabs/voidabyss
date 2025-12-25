package editor

// toggleFileTree opens or closes the file tree panel
func (e *Editor) toggleFileTree() {
	if e.treeOpen {
		e.treeOpen = false
		e.focusTree = false
		e.statusMsg = "file tree closed"
	} else {
		// Create file tree if not exists
		if e.fileTree == nil {
			// Use the directory of the current file or current working directory
			rootPath := "."
			if e.filename != "" {
				rootPath = e.filename
			}

			ft, err := NewFileTree(rootPath)
			if err != nil {
				e.statusMsg = "error opening tree: " + err.Error()
				return
			}
			e.fileTree = ft
			e.treePanelWidth = 30
		}
		e.treeOpen = true
		e.focusTree = true
		e.statusMsg = "file tree opened"
	}
}

// handleTreeInput handles input when the file tree has focus
func (e *Editor) handleTreeInput(key rune) {
	switch key {
	case 'j':
		e.fileTree.moveDown()
	case 'k':
		e.fileTree.moveUp()
	case 'c':
		// Close/collapse current directory or parent if on a file
		node := e.fileTree.getCurrentNode()
		if node != nil {
			if node.isDir && node.expanded {
				// Current node is an expanded directory, collapse it
				node.expanded = false
				e.fileTree.rebuildFlat()
				e.fileTree.ensureCursorValid()
				e.statusMsg = "collapsed " + node.name
			} else if !node.isDir && node.parent != nil {
				// Current node is a file, collapse its parent
				parent := node.parent
				if parent.expanded {
					parent.expanded = false
					e.fileTree.rebuildFlat()
					e.fileTree.ensureCursorValid()
					e.statusMsg = "collapsed " + parent.name
				}
			}
		}
	case '\r', '\n': // Enter key
		node := e.fileTree.getCurrentNode()
		if node != nil {
			if node.isDir {
				// Expand directory
				if !node.expanded {
					node.expanded = true
					e.fileTree.rebuildFlat()
					e.fileTree.ensureCursorValid()
					e.statusMsg = "expanded " + node.name
				} else {
					// Already expanded, collapse it
					node.expanded = false
					e.fileTree.rebuildFlat()
					e.fileTree.ensureCursorValid()
					e.statusMsg = "collapsed " + node.name
				}
			} else {
				// Open file in buffer
				e.openFile(node.path)
				e.focusTree = false
				e.statusMsg = "opened " + node.name
			}
		}
	case 'r':
		// TODO: Implement rename
		e.statusMsg = "rename not yet implemented"
	case 'q':
		e.treeOpen = false
		e.focusTree = false
	}
}
