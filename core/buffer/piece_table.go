package buffer

import (
	"errors"
)

// Source indicates whether a piece references the original file or the add buffer.
type Source uint8

const (
	SrcOriginal Source = iota
	SrcAdd
)

type piece struct {
	src   Source
	start int // rune index into the backing store (original/add)
	len   int // rune length
}

type Buffer struct {
	original []rune
	add      []rune
	pieces   []piece

	undo []op
	redo []op
}

// NewFromString creates a buffer where the initial contents live in "original".
func NewFromString(s string) *Buffer {
	r := []rune(s)
	b := &Buffer{
		original: r,
		add:      nil,
		pieces:   nil,
	}
	if len(r) == 0 {
		b.pieces = []piece{} // empty
	} else {
		b.pieces = []piece{{src: SrcOriginal, start: 0, len: len(r)}}
	}
	return b
}

func (b *Buffer) Len() int {
	total := 0
	for _, p := range b.pieces {
		total += p.len
	}
	return total
}

func (b *Buffer) String() string {
	out := make([]rune, 0, b.Len())
	for _, p := range b.pieces {
		out = append(out, b.readPiece(p)...)
	}
	return string(out)
}

// Slice returns the text in [start,end) (rune indices).
func (b *Buffer) Slice(start, end int) (string, error) {
	if start < 0 || end < start || end > b.Len() {
		return "", errors.New("slice: out of bounds")
	}
	if start == end {
		return "", nil
	}
	out := make([]rune, 0, end-start)
	cur := 0
	for _, p := range b.pieces {
		if p.len == 0 {
			continue
		}
		pStart := cur
		pEnd := cur + p.len
		cur = pEnd

		// no overlap
		if end <= pStart || start >= pEnd {
			continue
		}

		// overlap
		lo := max(start, pStart)
		hi := min(end, pEnd)
		offset := lo - pStart
		length := hi - lo

		chunk := b.readPiece(piece{src: p.src, start: p.start + offset, len: length})
		out = append(out, chunk...)
	}
	return string(out), nil
}

// Insert inserts text at position pos (rune index). Records undo.
func (b *Buffer) Insert(pos int, text string) error {
	if pos < 0 || pos > b.Len() {
		return errors.New("insert: out of bounds")
	}
	if text == "" {
		return nil
	}
	op := insertOp{pos: pos, text: []rune(text)}
	inv, err := b.apply(op)
	if err != nil {
		return err
	}
	b.undo = append(b.undo, inv)
	b.redo = b.redo[:0]
	return nil
}

// Delete deletes range [start,end) (rune indices). Records undo.
func (b *Buffer) Delete(start, end int) error {
	if start < 0 || end < start || end > b.Len() {
		return errors.New("delete: out of bounds")
	}
	if start == end {
		return nil
	}
	// capture deleted text for undo
	deleted, err := b.Slice(start, end)
	if err != nil {
		return err
	}

	op := deleteOp{start: start, end: end}
	inv, err := b.apply(op)
	if err != nil {
		return err
	}

	// our inverse needs the deleted payload to restore precisely
	// (inv is insertOp but with empty text; fill it)
	if ins, ok := inv.(insertOp); ok {
		ins.text = []rune(deleted)
		inv = ins
	}

	b.undo = append(b.undo, inv)
	b.redo = b.redo[:0]
	return nil
}

func (b *Buffer) Undo() bool {
	if len(b.undo) == 0 {
		return false
	}
	last := b.undo[len(b.undo)-1]
	b.undo = b.undo[:len(b.undo)-1]

	inv, err := b.apply(last)
	if err != nil {
		// if this happens, your internal invariants are broken
		return false
	}
	b.redo = append(b.redo, inv)
	return true
}

func (b *Buffer) Redo() bool {
	if len(b.redo) == 0 {
		return false
	}
	last := b.redo[len(b.redo)-1]
	b.redo = b.redo[:len(b.redo)-1]

	inv, err := b.apply(last)
	if err != nil {
		return false
	}
	b.undo = append(b.undo, inv)
	return true
}

// ----- ops -----

type op interface {
	applyTo(b *Buffer) (inverse op, err error)
}

func (b *Buffer) apply(o op) (op, error) {
	return o.applyTo(b)
}

type insertOp struct {
	pos  int
	text []rune
}

func (o insertOp) applyTo(b *Buffer) (op, error) {
	if o.pos < 0 || o.pos > b.Len() {
		return nil, errors.New("insertOp: out of bounds")
	}
	if len(o.text) == 0 {
		return deleteOp{start: o.pos, end: o.pos}, nil
	}

	// append to add buffer
	addStart := len(b.add)
	b.add = append(b.add, o.text...)
	newPiece := piece{src: SrcAdd, start: addStart, len: len(o.text)}

	left, right := splitPiecesAt(b.pieces, o.pos)
	merged := append([]piece{}, left...)
	merged = append(merged, newPiece)
	merged = append(merged, right...)
	b.pieces = coalesce(merged)

	// inverse is delete of inserted range
	return deleteOp{start: o.pos, end: o.pos + len(o.text)}, nil
}

type deleteOp struct {
	start int
	end   int
}

func (o deleteOp) applyTo(b *Buffer) (op, error) {
	if o.start < 0 || o.end < o.start || o.end > b.Len() {
		return nil, errors.New("deleteOp: out of bounds")
	}
	if o.start == o.end {
		return insertOp{pos: o.start, text: nil}, nil
	}

	// remove [start,end) by splitting at start and end, then discarding middle
	left, midRight := splitPiecesAt(b.pieces, o.start)
	_, right := splitPiecesAt(midRight, o.end-o.start)
	b.pieces = coalesce(append(left, right...))

	// inverse is insert, but caller supplies deleted payload for accurate undo
	return insertOp{pos: o.start, text: nil}, nil
}

// ----- piece table internals -----

func (b *Buffer) readPiece(p piece) []rune {
	if p.len <= 0 {
		return nil
	}
	switch p.src {
	case SrcOriginal:
		return b.original[p.start : p.start+p.len]
	case SrcAdd:
		return b.add[p.start : p.start+p.len]
	default:
		return nil
	}
}

// splitPiecesAt splits pieces into left/right such that left has exactly `pos` runes.
func splitPiecesAt(pieces []piece, pos int) (left []piece, right []piece) {
	if pos <= 0 {
		return []piece{}, append([]piece{}, pieces...)
	}

	cur := 0
	for i := 0; i < len(pieces); i++ {
		p := pieces[i]
		if p.len == 0 {
			continue
		}
		next := cur + p.len

		// pos lies after this piece
		if pos >= next {
			left = append(left, p)
			cur = next
			continue
		}

		// pos lies within this piece -> split
		within := pos - cur
		if within <= 0 {
			// split at piece boundary (before p)
			right = append(right, pieces[i:]...)
			return coalesce(left), coalesce(right)
		}

		pLeft := piece{src: p.src, start: p.start, len: within}
		pRight := piece{src: p.src, start: p.start + within, len: p.len - within}

		left = append(left, pLeft)
		right = append(right, pRight)
		right = append(right, pieces[i+1:]...)
		return coalesce(left), coalesce(right)
	}

	// pos beyond end -> all in left
	return coalesce(left), []piece{}
}

func coalesce(pieces []piece) []piece {
	out := make([]piece, 0, len(pieces))
	for _, p := range pieces {
		if p.len <= 0 {
			continue
		}
		n := len(out)
		if n == 0 {
			out = append(out, p)
			continue
		}
		prev := out[n-1]
		// merge adjacent pieces from same source and contiguous backing region
		if prev.src == p.src && prev.start+prev.len == p.start {
			out[n-1] = piece{src: prev.src, start: prev.start, len: prev.len + p.len}
			continue
		}
		out = append(out, p)
	}
	return out
}
