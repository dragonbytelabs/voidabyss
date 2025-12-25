package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewFileTree(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644)

	ft, err := NewFileTree(tmpDir)
	if err != nil {
		t.Fatalf("NewFileTree failed: %v", err)
	}

	if ft.root == nil {
		t.Fatal("root should not be nil")
	}

	if !ft.root.isDir {
		t.Error("root should be a directory")
	}

	if len(ft.flat) == 0 {
		t.Error("flat list should be populated")
	}
}

func TestFileTreeIgnore(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, ".git"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)

	ft, err := NewFileTree(tmpDir)
	if err != nil {
		t.Fatalf("NewFileTree failed: %v", err)
	}

	for _, child := range ft.root.children {
		if child.name == ".git" || child.name == "node_modules" {
			t.Errorf("ignored item %s should not be in tree", child.name)
		}
	}

	found := false
	for _, child := range ft.root.children {
		if child.name == "src" {
			found = true
		}
	}
	if !found {
		t.Error("src should be in tree")
	}
}

func TestFileTreeNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("test"), 0644)

	ft, err := NewFileTree(tmpDir)
	if err != nil {
		t.Fatalf("NewFileTree failed: %v", err)
	}

	start := ft.cursor
	ft.moveDown()
	if ft.cursor <= start {
		t.Error("cursor should move down")
	}

	ft.moveUp()
	if ft.cursor != start {
		t.Error("cursor should move back up")
	}
}

func TestFileTreeExpand(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "dir1"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "dir1", "file.txt"), []byte("test"), 0644)

	ft, err := NewFileTree(tmpDir)
	if err != nil {
		t.Fatalf("NewFileTree failed: %v", err)
	}

	var dirNode *FileTreeNode
	for _, child := range ft.root.children {
		if child.isDir && child.name == "dir1" {
			dirNode = child
			break
		}
	}

	if dirNode == nil {
		t.Skip("dir1 not found")
	}

	dirNode.expanded = true
	ft.rebuildFlat()
	expanded := len(ft.flat)

	dirNode.expanded = false
	ft.rebuildFlat()
	collapsed := len(ft.flat)

	if collapsed >= expanded {
		t.Error("collapsed should have fewer items")
	}
}

func TestFileTreeDisplayLines(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644)

	ft, err := NewFileTree(tmpDir)
	if err != nil {
		t.Fatalf("NewFileTree failed: %v", err)
	}

	lines := ft.getDisplayLines()
	if len(lines) == 0 {
		t.Error("should have display lines")
	}

	if !strings.HasPrefix(lines[0], ">") {
		t.Error("first line should have cursor marker")
	}
}
