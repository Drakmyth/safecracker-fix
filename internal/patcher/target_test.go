package patcher

import (
	"crypto/sha256"
	"testing"
)

func TestTargetValidate(t *testing.T) {
	const (
		DEFAULT_TARGET_ID   TargetID = "test-target"
		DEFAULT_TARGET_SIZE          = 10
		PATCH_A             PatchID  = "a"
		PATCH_B             PatchID  = "b"
	)

	edit := func(offset int64, length int) Edit {
		original := make([]byte, length)
		replacement := make([]byte, length)

		for i := range replacement {
			replacement[i] = 0xFF
		}

		return Edit{
			Offset:      offset,
			Original:    original,
			Replacement: replacement,
		}
	}

	patch := func(id PatchID, edits ...Edit) Patch {
		return Patch{
			ID:    id,
			Edits: edits,
		}
	}

	target := func(id TargetID, size int64, patches ...Patch) Target {
		return Target{
			ID:           id,
			OriginalHash: sha256.Sum256(make([]byte, DEFAULT_TARGET_SIZE)),
			Size:         size,
			Patches:      patches,
		}
	}

	defaultTarget := func(patches ...Patch) Target {
		return target(DEFAULT_TARGET_ID, DEFAULT_TARGET_SIZE, patches...)
	}

	tests := []struct {
		name    string
		target  Target
		wantErr bool
	}{
		{
			name: "valid target",
			target: defaultTarget(
				patch(PATCH_A, edit(0, 2)),
				patch(PATCH_B, edit(5, 2)),
			),
		},
		{
			name: "empty ID",
			target: target("", DEFAULT_TARGET_SIZE,
				patch(PATCH_A, edit(0, 2)),
			),
			wantErr: true,
		},
		{
			name: "zero size",
			target: target(DEFAULT_TARGET_ID, 0,
				patch(PATCH_A, edit(0, 2)),
			),
			wantErr: true,
		},
		{
			name: "negative size",
			target: target(DEFAULT_TARGET_ID, -1,
				patch(PATCH_A, edit(0, 2)),
			),
			wantErr: true,
		},
		{
			name:    "no patches",
			target:  defaultTarget(),
			wantErr: true,
		},
		{
			name: "invalid patch",
			target: defaultTarget(
				patch(PATCH_A),
			),
			wantErr: true,
		},
		{
			name: "duplicate patch IDs",
			target: defaultTarget(
				patch(PATCH_A, edit(0, 2)),
				patch(PATCH_A, edit(5, 2)),
			),
			wantErr: true,
		},
		{
			name: "adjacent patches",
			target: defaultTarget(
				patch(PATCH_A, edit(0, 2)),
				patch(PATCH_B, edit(2, 2)),
			),
		},
		{
			name: "overlapping patches",
			target: defaultTarget(
				patch(PATCH_A, edit(0, 2)),
				patch(PATCH_B, edit(1, 2)),
			),
			wantErr: true,
		},
		{
			name: "identical patch ranges",
			target: defaultTarget(
				patch(PATCH_A, edit(0, 2)),
				patch(PATCH_B, edit(0, 2)),
			),
			wantErr: true,
		},
		{
			name: "patch exceeds target size",
			target: defaultTarget(
				patch(PATCH_A, edit(9, 2)),
			),
			wantErr: true,
		},
		{
			name: "overlap in secondary edit",
			target: defaultTarget(
				patch(PATCH_A, edit(0, 2), edit(6, 1)),
				patch(PATCH_B, edit(5, 2)),
			),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.target.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() = %v, wantErr %v",
					err, tt.wantErr)
			}
		})
	}
}
