package buffer

import "testing"

func TestUndoGrouping(t *testing.T) {
	b := NewFromString("")

	// Begin a group
	b.BeginUndoGroup()

	// Insert multiple characters
	_ = b.Insert(0, "h")
	_ = b.Insert(1, "e")
	_ = b.Insert(2, "l")
	_ = b.Insert(3, "l")
	_ = b.Insert(4, "o")

	// End group
	b.EndUndoGroup()

	if got := b.String(); got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}

	// Single undo should remove entire group
	if !b.Undo() {
		t.Fatal("undo should succeed")
	}

	if got := b.String(); got != "" {
		t.Fatalf("after undo, expected empty, got %q", got)
	}

	// Redo should restore entire group
	if !b.Redo() {
		t.Fatal("redo should succeed")
	}

	if got := b.String(); got != "hello" {
		t.Fatalf("after redo, expected 'hello', got %q", got)
	}
}

func TestUndoGroupingWithDelete(t *testing.T) {
	b := NewFromString("abc")

	b.BeginUndoGroup()
	_ = b.Delete(0, 1) // delete 'a'
	_ = b.Insert(0, "x")
	_ = b.Insert(1, "y")
	b.EndUndoGroup()

	if got := b.String(); got != "xybc" {
		t.Fatalf("expected 'xybc', got %q", got)
	}

	// Single undo restores original
	if !b.Undo() {
		t.Fatal("undo should succeed")
	}

	if got := b.String(); got != "abc" {
		t.Fatalf("after undo, expected 'abc', got %q", got)
	}

	// Redo reapplies group
	if !b.Redo() {
		t.Fatal("redo should succeed")
	}

	if got := b.String(); got != "xybc" {
		t.Fatalf("after redo, expected 'xybc', got %q", got)
	}
}

func TestEmptyGroup(t *testing.T) {
	b := NewFromString("test")

	b.BeginUndoGroup()
	// No operations
	b.EndUndoGroup()

	// Buffer should still be "test"
	if got := b.String(); got != "test" {
		t.Fatalf("expected 'test', got %q", got)
	}

	// Empty group should not have been added to undo stack
	if b.Undo() {
		t.Fatal("undo should not succeed for empty group")
	}

	// Buffer should still be "test" since undo did nothing
	if got := b.String(); got != "test" {
		t.Fatalf("after failed undo, expected 'test', got %q", got)
	}
}
