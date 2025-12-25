package editor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestToggleFileTree(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	ed := newTestEditor(t, "")
	ed.filename = testFile

	if ed.treeOpen {
		t.Error("tree should start closed")
	}

	ed.toggleFileTree()
	if !ed.treeOpen {
		t.Error("tree should be open")
	}
	if ed.fileTree == nil {
		t.Error("fileTree should be created")
	}

	ed.toggleFileTree()
	if ed.treeOpen {
		t.Error("tree should be closed")
	}
}

func TestTreeNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)

	ed := newTestEditor(t, "")
	ed.filename = filepath.Join(tmpDir, "file1.txt")
	ed.toggleFileTree()

	start := ed.fileTree.cursor
	ed.handleTreeInput('j')
	if ed.fileTree.cursor <= start {
		t.Error("j should move cursor down")
	}

	ed.handleTreeInput('k')
	if ed.fileTree.cursor >= start+1 {
		t.Error("k should move cursor up")
	}
}

func TestTreeOpenFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("content"), 0644)

	ed := newTestEditor(t, "")
	ed.filename = tmpDir
	ed.toggleFileTree()

	var fileNode *FileTreeNode
	for i, node := range ed.fileTree.flat {
		if !node.isDir && node.name == "test.txt" {
			fileNode = node
			ed.fileTree.cursor = i
			break
		}
	}

	if fileNode == nil {
		t.Skip("test.txt not found in tree")
	}

	ed.handleTreeInput('\n')

	if ed.focusTree {
		t.Error("should lose tree focus after opening file")
	}
}

func TestTreeQuit(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)

	ed := newTestEditor(t, "")
	ed.filename = filepath.Join(tmpDir, "test.txt")
	ed.toggleFileTree()

	if !ed.treeOpen {
		t.Fatal("tree should be open")
	}

	ed.handleTreeInput('q')

	if ed.treeOpen {
		t.Error("tree should be closed after q")
	}
}
