package editor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenFile(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "test1.txt")
	file2 := filepath.Join(tmpDir, "test2.txt")

	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)

	ed := newTestEditor(t, "")

	ed.openFile(file1)
	if len(ed.buffers) != 1 {
		t.Errorf("expected 1 buffer, got %d", len(ed.buffers))
	}

	ed.openFile(file2)
	if len(ed.buffers) != 2 {
		t.Errorf("expected 2 buffers, got %d", len(ed.buffers))
	}

	ed.openFile(file1)
	if len(ed.buffers) != 2 {
		t.Errorf("expected 2 buffers after reopening, got %d", len(ed.buffers))
	}
}

func TestBufferSwitching(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "test1.txt")
	file2 := filepath.Join(tmpDir, "test2.txt")
	file3 := filepath.Join(tmpDir, "test3.txt")

	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)
	os.WriteFile(file3, []byte("content3"), 0644)

	ed := newTestEditor(t, "")
	ed.openFile(file1)
	ed.openFile(file2)
	ed.openFile(file3)

	if ed.currentBuffer != 2 {
		t.Errorf("expected currentBuffer 2, got %d", ed.currentBuffer)
	}

	ed.nextBuffer()
	if ed.currentBuffer != 0 {
		t.Errorf("expected wrap to buffer 0, got %d", ed.currentBuffer)
	}

	ed.prevBuffer()
	if ed.currentBuffer != 2 {
		t.Errorf("expected wrap to buffer 2, got %d", ed.currentBuffer)
	}
}

func TestBufferDirtyTracking(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "test1.txt")
	file2 := filepath.Join(tmpDir, "test2.txt")

	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)

	ed := newTestEditor(t, "")
	ed.openFile(file1)
	ed.openFile(file2)

	ed.dirty = true
	ed.syncToBuffer()

	ed.prevBuffer()
	if ed.dirty {
		t.Error("file1 should not be dirty")
	}

	ed.nextBuffer()
	if !ed.dirty {
		t.Error("file2 should be dirty")
	}
}
