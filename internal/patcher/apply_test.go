package patcher

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

func TestTargetApplyAndReverse(t *testing.T) {
	const (
		DEFAULT_TARGET_ID   TargetID = "test-target"
		DEFAULT_TARGET_SIZE          = 10
		PATCH_A             PatchID  = "a"
		PATCH_B             PatchID  = "b"
	)

	original := make([]byte, DEFAULT_TARGET_SIZE)

	target := Target{
		ID:           DEFAULT_TARGET_ID,
		Size:         DEFAULT_TARGET_SIZE,
		OriginalHash: sha256.Sum256(original),
		Patches: []Patch{
			{
				ID: PATCH_A,
				Edits: []Edit{
					{
						Offset:      0,
						Original:    []byte{0x00},
						Replacement: []byte{0xAA},
					},
					{
						Offset:      2,
						Original:    []byte{0x00},
						Replacement: []byte{0xAA},
					},
				},
			},
			{
				ID: PATCH_B,
				Edits: []Edit{
					{
						Offset:      5,
						Original:    []byte{0x00},
						Replacement: []byte{0xBB},
					},
				},
			},
		},
	}

	binary := func(edits map[int]byte) []byte {
		data := bytes.Clone(original)

		for offset, value := range edits {
			data[offset] = value
		}

		return data
	}

	tests := []struct {
		name      string
		operation func([]byte, PatchID) ([]byte, error)
		id        PatchID
		data      []byte
		want      []byte
		wantErr   bool
	}{
		{
			name:      "apply to original",
			operation: target.Apply,
			id:        PATCH_A,
			data:      binary(nil),
			want:      binary(map[int]byte{0: 0xAA, 2: 0xAA}),
		},
		{
			name:      "apply already applied patch",
			operation: target.Apply,
			id:        PATCH_A,
			data:      binary(map[int]byte{0: 0xAA, 2: 0xAA}),
			want:      binary(map[int]byte{0: 0xAA, 2: 0xAA}),
		},
		{
			name:      "complete partial patch",
			operation: target.Apply,
			id:        PATCH_A,
			data:      binary(map[int]byte{0: 0xAA}),
			want:      binary(map[int]byte{0: 0xAA, 2: 0xAA}),
		},
		{
			name:      "apply preserves other patch",
			operation: target.Apply,
			id:        PATCH_A,
			data:      binary(map[int]byte{5: 0xBB}),
			want: binary(map[int]byte{
				0: 0xAA,
				2: 0xAA,
				5: 0xBB,
			}),
		},
		{
			name:      "reverse applied patch",
			operation: target.Reverse,
			id:        PATCH_A,
			data:      binary(map[int]byte{0: 0xAA, 2: 0xAA}),
			want:      binary(nil),
		},
		{
			name:      "reverse unapplied patch",
			operation: target.Reverse,
			id:        PATCH_A,
			data:      binary(nil),
			want:      binary(nil),
		},
		{
			name:      "reverse partial patch",
			operation: target.Reverse,
			id:        PATCH_A,
			data:      binary(map[int]byte{2: 0xAA}),
			want:      binary(nil),
		},
		{
			name:      "reverse preserves other patch",
			operation: target.Reverse,
			id:        PATCH_A,
			data: binary(map[int]byte{
				0: 0xAA,
				2: 0xAA,
				5: 0xBB,
			}),
			want: binary(map[int]byte{5: 0xBB}),
		},
		{
			name:      "apply unknown patch",
			operation: target.Apply,
			id:        "unknown",
			data:      binary(nil),
			wantErr:   true,
		},
		{
			name:      "reverse unknown patch",
			operation: target.Reverse,
			id:        "unknown",
			data:      binary(nil),
			wantErr:   true,
		},
		{
			name:      "apply to unexpected edit bytes",
			operation: target.Apply,
			id:        PATCH_A,
			data:      binary(map[int]byte{0: 0xCC}),
			wantErr:   true,
		},
		{
			name:      "reverse with unrelated modification",
			operation: target.Reverse,
			id:        PATCH_A,
			data: binary(map[int]byte{
				0: 0xAA,
				2: 0xAA,
				9: 0xCC,
			}),
			wantErr: true,
		},
		{
			name:      "apply to wrong binary size",
			operation: target.Apply,
			id:        PATCH_A,
			data:      make([]byte, DEFAULT_TARGET_SIZE-1),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := bytes.Clone(tt.data)

			result, err := tt.operation(tt.data, tt.id)

			if !bytes.Equal(tt.data, before) {
				t.Error("operation modified input data")
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				if result != nil {
					t.Error("operation returned data on error")
				}
				return
			}

			if !bytes.Equal(result, tt.want) {
				t.Errorf("result = %x, want %x", result, tt.want)
			}

			// Verify that the result does not share storage
			// with the input, even for idempotent operations.
			if len(result) > 0 {
				result[0] ^= 0xFF

				if !bytes.Equal(tt.data, before) {
					t.Error("result aliases input data")
				}
			}
		})
	}
}

func TestTargetPatchComposition(t *testing.T) {
	const (
		PATCH_A PatchID = "a"
		PATCH_B PatchID = "b"
	)

	original := []byte{0x00, 0x00, 0x00, 0x00}

	target := Target{
		ID:           "composition-test",
		Size:         int64(len(original)),
		OriginalHash: sha256.Sum256(original),
		Patches: []Patch{
			{
				ID: PATCH_A,
				Edits: []Edit{
					{
						Offset:      0,
						Original:    []byte{0x00},
						Replacement: []byte{0xAA},
					},
				},
			},
			{
				ID: PATCH_B,
				Edits: []Edit{
					{
						Offset:      2,
						Original:    []byte{0x00},
						Replacement: []byte{0xBB},
					},
				},
			},
		},
	}

	apply := func(data []byte, id PatchID) []byte {
		t.Helper()

		result, err := target.Apply(data, id)
		if err != nil {
			t.Fatalf("Apply(%q): %v", id, err)
		}

		return result
	}

	reverse := func(data []byte, id PatchID) []byte {
		t.Helper()

		result, err := target.Reverse(data, id)
		if err != nil {
			t.Fatalf("Reverse(%q): %v", id, err)
		}

		return result
	}

	// Applying patches in either order produces the same binary.
	ab := apply(apply(original, PATCH_A), PATCH_B)
	ba := apply(apply(original, PATCH_B), PATCH_A)

	if !bytes.Equal(ab, ba) {
		t.Errorf("patch order matters: AB = %x, BA = %x", ab, ba)
	}

	want := []byte{0xAA, 0x00, 0xBB, 0x00}

	if !bytes.Equal(ab, want) {
		t.Errorf("combined result = %x, want %x", ab, want)
	}

	// Reversing patches in either order restores the original.
	reversedAB := reverse(reverse(ab, PATCH_A), PATCH_B)
	reversedBA := reverse(reverse(ba, PATCH_B), PATCH_A)

	if !bytes.Equal(reversedAB, original) {
		t.Errorf("reverse AB = %x, want %x", reversedAB, original)
	}

	if !bytes.Equal(reversedBA, original) {
		t.Errorf("reverse BA = %x, want %x", reversedBA, original)
	}
}
