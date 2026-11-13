package patcher

import (
	"bytes"
	"fmt"
)

// Edit describes a fixed-length byte replacement within a binary
type Edit struct {
	Offset      int64
	Original    []byte
	Replacement []byte
}

type PatchID string

// Patch describes a complete, reversible functional change to a binary
type Patch struct {
	ID    PatchID
	Name  string
	Edits []Edit
}

func (e Edit) Validate(size int64) error {
	if e.Offset < 0 {
		return fmt.Errorf("negative edit offset: %d", e.Offset)
	}

	if len(e.Original) == 0 {
		return fmt.Errorf("edit at offset %#x is empty", e.Offset)
	}

	if len(e.Original) != len(e.Replacement) {
		return fmt.Errorf("edit at offset %#x has mismatched byte lengths", e.Offset)
	}

	if bytes.Equal(e.Original, e.Replacement) {
		return fmt.Errorf("edit at offset %#x does not change any bytes", e.Offset)
	}

	if e.Offset > size || int64(len(e.Original)) > size-e.Offset {
		return fmt.Errorf("edit at offset %#x exceeds binary size", e.Offset)
	}

	return nil
}

func (p Patch) Validate(size int64) error {
	if p.ID == "" {
		return fmt.Errorf("patch ID is empty")
	}

	if len(p.Edits) == 0 {
		return fmt.Errorf("patch %q contains no edits", p.ID)
	}

	for i, edit := range p.Edits {
		if err := edit.Validate(size); err != nil {
			return fmt.Errorf("patch %q, edit %d: %w", p.ID, i, err)
		}
	}

	for i := range p.Edits {
		for j := i + 1; j < len(p.Edits); j++ {
			if editsOverlap(p.Edits[i], p.Edits[j]) {
				return fmt.Errorf("patch %q: edits %d and %d overlap", p.ID, i, j)
			}
		}
	}

	return nil
}

// editsOverlap reports whether two validated edits modify and of the same bytes.
func editsOverlap(a, b Edit) bool {
	aEnd := a.Offset + int64(len(a.Original))
	bEnd := b.Offset + int64(len(b.Original))
	return a.Offset < bEnd && b.Offset < aEnd
}
