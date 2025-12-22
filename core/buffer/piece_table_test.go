package buffer

import "testing"

func TestInsertAndString(t *testing.T) {
	b := NewFromString("hello")
	if got := b.String(); got != "hello" {
		t.Fatalf("expected hello, got %q", got)
	}

	if err := b.Insert(5, " world"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if got := b.String(); got != "hello world" {
		t.Fatalf("expected hello world, got %q", got)
	}

	if err := b.Insert(0, ">>"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if got := b.String(); got != ">>hello world" {
		t.Fatalf("expected >>hello world, got %q", got)
	}
}

func TestDeleteRange(t *testing.T) {
	b := NewFromString("abc def ghi")
	// delete "def "
	if err := b.Delete(4, 8); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if got := b.String(); got != "abc ghi" {
		t.Fatalf("expected %q, got %q", "abc ghi", got)
	}
}

func TestUndoRedo(t *testing.T) {
	b := NewFromString("one two three")

	if err := b.Delete(4, 8); err != nil { // delete "two "
		t.Fatalf("delete: %v", err)
	}
	if got := b.String(); got != "one three" {
		t.Fatalf("after delete: got %q", got)
	}

	if !b.Undo() {
		t.Fatalf("expected undo true")
	}
	if got := b.String(); got != "one two three" {
		t.Fatalf("after undo: got %q", got)
	}

	if !b.Redo() {
		t.Fatalf("expected redo true")
	}
	if got := b.String(); got != "one three" {
		t.Fatalf("after redo: got %q", got)
	}
}

func TestUnicodeRunes(t *testing.T) {
	b := NewFromString("a🙂c")
	if b.Len() != 3 {
		t.Fatalf("expected len 3 runes, got %d", b.Len())
	}
	if err := b.Insert(1, "β"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	if got := b.String(); got != "aβ🙂c" {
		t.Fatalf("expected %q, got %q", "aβ🙂c", got)
	}
}

func TestSlice(t *testing.T) {
	b := NewFromString("0123456789")
	s, err := b.Slice(3, 7)
	if err != nil {
		t.Fatalf("slice: %v", err)
	}
	if s != "3456" {
		t.Fatalf("expected 3456, got %q", s)
	}
}
