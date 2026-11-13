package patcher

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// PatchState describes the installation state of a patch.
type PatchState int

const (
	PATCH_STATE_UNKNOWN PatchState = iota
	PATCH_STATE_UNAPPLIED
	PATCH_STATE_APPLIED
	PATCH_STATE_PARTIAL
)

func (ps PatchState) String() string {
	switch ps {
	case PATCH_STATE_UNKNOWN:
		return "Unknown"
	case PATCH_STATE_UNAPPLIED:
		return "Unapplied"
	case PATCH_STATE_APPLIED:
		return "Applied"
	case PATCH_STATE_PARTIAL:
		return "Partial"
	default:
		return fmt.Sprintf("PatchState(%d)", int(ps))
	}
}

// Inspection describes the state of a recognized binary.
type Inspection struct {
	TargetID TargetID
	Patches  map[PatchID]PatchState
}

// Inspect verifies that data matches the target and determines the installation state of each patch.
// The input data is never modified.
func (t Target) Inspect(data []byte) (*Inspection, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}

	if int64(len(data)) != t.Size {
		return nil, fmt.Errorf("target %q: expected %d bytes, got %d", t.ID, t.Size, len(data))
	}

	normalized := bytes.Clone(data)

	result := &Inspection{
		TargetID: t.ID,
		Patches:  make(map[PatchID]PatchState, len(t.Patches)),
	}

	for _, patch := range t.Patches {
		applied := 0

		for _, edit := range patch.Edits {
			start := edit.Offset
			end := start + int64(len(edit.Original))

			current := data[start:end]

			switch {
			case bytes.Equal(current, edit.Original):
				// This edit has not been applied.
			case bytes.Equal(current, edit.Replacement):
				applied++

				// Reverse the edit in our temporary copy.
				copy(normalized[start:end], edit.Original)
			default:
				return nil, fmt.Errorf("patch %q: unexpected bytes at offset %#x", patch.ID, start)
			}
		}

		switch applied {
		case 0:
			result.Patches[patch.ID] = PATCH_STATE_UNAPPLIED
		case len(patch.Edits):
			result.Patches[patch.ID] = PATCH_STATE_APPLIED
		default:
			result.Patches[patch.ID] = PATCH_STATE_PARTIAL
		}
	}

	if sha256.Sum256(normalized) != t.OriginalHash {
		return nil, fmt.Errorf("target %q: original hash mismatch", t.ID)
	}

	return result, nil
}
