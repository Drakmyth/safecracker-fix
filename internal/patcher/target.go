package patcher

import (
	"crypto/sha256"
	"fmt"
)

type TargetID string

// Target describes a supported binary version and the patches that can be applied to it
type Target struct {
	ID           TargetID
	Name         string
	OriginalHash [sha256.Size]byte
	Size         int64
	Patches      []Patch
}

// Validate checks that the target and its patches are well-formed and independently composable.
func (t Target) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("target ID is empty")
	}

	if t.Size <= 0 {
		return fmt.Errorf("target %q has invalid size: %d", t.ID, t.Size)
	}

	if len(t.Patches) == 0 {
		return fmt.Errorf("target %q contains no patches", t.ID)
	}

	seen := make(map[PatchID]struct{}, len(t.Patches))

	for _, patch := range t.Patches {
		if err := patch.Validate(t.Size); err != nil {
			return fmt.Errorf("target %q: %w", t.ID, err)
		}

		if _, exists := seen[patch.ID]; exists {
			return fmt.Errorf("target %q contains duplicate patch ID %q", t.ID, patch.ID)
		}

		seen[patch.ID] = struct{}{}
	}

	for i := range t.Patches {
		for j := i + 1; j < len(t.Patches); j++ {
			a := t.Patches[i]
			b := t.Patches[j]

			for ai, ae := range a.Edits {
				for bi, be := range b.Edits {
					if editsOverlap(ae, be) {
						return fmt.Errorf("target %q: patch %q edit %d overlaps patch %q edit %d", t.ID, a.ID, ai, b.ID, bi)
					}
				}
			}
		}
	}

	return nil
}
