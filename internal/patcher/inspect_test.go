package patcher

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

func TestTargetInspect(t *testing.T) {
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
		name    string
		data    []byte
		stateA  PatchState
		stateB  PatchState
		wantErr bool
	}{
		{
			name:   "original binary",
			data:   binary(nil),
			stateA: PATCH_STATE_UNAPPLIED,
			stateB: PATCH_STATE_UNAPPLIED,
		},
		{
			name:   "patch A applied",
			data:   binary(map[int]byte{0: 0xAA, 2: 0xAA}),
			stateA: PATCH_STATE_APPLIED,
			stateB: PATCH_STATE_UNAPPLIED,
		},
		{
			name:   "patch B applied",
			data:   binary(map[int]byte{5: 0xBB}),
			stateA: PATCH_STATE_UNAPPLIED,
			stateB: PATCH_STATE_APPLIED,
		},
		{
			name: "both patches applied",
			data: binary(map[int]byte{
				0: 0xAA,
				2: 0xAA,
				5: 0xBB,
			}),
			stateA: PATCH_STATE_APPLIED,
			stateB: PATCH_STATE_APPLIED,
		},
		{
			name:   "patch A partially applied",
			data:   binary(map[int]byte{0: 0xAA}),
			stateA: PATCH_STATE_PARTIAL,
			stateB: PATCH_STATE_UNAPPLIED,
		},
		{
			name:   "patch A partial and patch B applied",
			data:   binary(map[int]byte{2: 0xAA, 5: 0xBB}),
			stateA: PATCH_STATE_PARTIAL,
			stateB: PATCH_STATE_APPLIED,
		},
		{
			name:    "unexpected edit bytes",
			data:    binary(map[int]byte{0: 0xCC}),
			wantErr: true,
		},
		{
			name:    "modification outside known edits",
			data:    binary(map[int]byte{9: 0xCC}),
			wantErr: true,
		},
		{
			name: "applied patch with unrelated modification",
			data: binary(map[int]byte{
				0: 0xAA,
				2: 0xAA,
				9: 0xCC,
			}),
			wantErr: true,
		},
		{
			name:    "binary too small",
			data:    make([]byte, DEFAULT_TARGET_SIZE-1),
			wantErr: true,
		},
		{
			name:    "binary too large",
			data:    make([]byte, DEFAULT_TARGET_SIZE+1),
			wantErr: true,
		},
		{
			name:    "empty binary",
			data:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := bytes.Clone(tt.data)

			result, err := target.Inspect(tt.data)

			if !bytes.Equal(tt.data, before) {
				t.Error("Inspect() modified input data")
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("Inspect() error = %v, wantErr %v",
					err, tt.wantErr)
			}

			if tt.wantErr {
				if result != nil {
					t.Error("Inspect() returned a result on error")
				}
				return
			}

			if result == nil {
				t.Fatal("Inspect() returned nil result")
			}

			if result.TargetID != target.ID {
				t.Errorf("TargetID = %q, want %q",
					result.TargetID, target.ID)
			}

			if len(result.Patches) != len(target.Patches) {
				t.Errorf("got %d patch states, want %d",
					len(result.Patches), len(target.Patches))
			}

			expected := map[PatchID]PatchState{
				PATCH_A: tt.stateA,
				PATCH_B: tt.stateB,
			}

			for id, want := range expected {
				got, exists := result.Patches[id]

				if !exists {
					t.Errorf("missing state for patch %q", id)
					continue
				}

				if got != want {
					t.Errorf("patch %q: state = %v, want %v",
						id, got, want)
				}
			}
		})
	}
}

func TestPatchStateString(t *testing.T) {
	tests := []struct {
		state PatchState
		want  string
	}{
		{PATCH_STATE_UNKNOWN, "Unknown"},
		{PATCH_STATE_UNAPPLIED, "Unapplied"},
		{PATCH_STATE_APPLIED, "Applied"},
		{PATCH_STATE_PARTIAL, "Partial"},
		{PatchState(999), "PatchState(999)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.state.String(); got != tt.want {
				t.Errorf("String() = %q, want %q",
					got, tt.want)
			}
		})
	}
}
