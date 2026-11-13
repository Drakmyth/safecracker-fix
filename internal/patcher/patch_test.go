package patcher

import "testing"

func TestEditValidate(t *testing.T) {
	const DEFAULT_SIZE int64 = 10

	tests := []struct {
		name    string
		edit    Edit
		size    int64
		wantErr bool
	}{
		{
			name: "valid edit",
			edit: Edit{2, []byte{0x20}, []byte{0x22}},
			size: DEFAULT_SIZE,
		},
		{
			name:    "negative offset",
			edit:    Edit{-1, []byte{0x20}, []byte{0x22}},
			size:    DEFAULT_SIZE,
			wantErr: true,
		},
		{
			name:    "empty edit",
			edit:    Edit{0, nil, nil},
			size:    DEFAULT_SIZE,
			wantErr: true,
		},
		{
			name:    "mismatched lengths",
			edit:    Edit{0, []byte{0x20}, []byte{0x22, 0x23}},
			size:    DEFAULT_SIZE,
			wantErr: true,
		},
		{
			name:    "identical bytes",
			edit:    Edit{0, []byte{0x20}, []byte{0x20}},
			size:    DEFAULT_SIZE,
			wantErr: true,
		},
		{
			name: "edit at last byte",
			edit: Edit{9, []byte{0x20}, []byte{0x22}},
			size: DEFAULT_SIZE,
		},
		{
			name:    "edit past end",
			edit:    Edit{DEFAULT_SIZE, []byte{0x20}, []byte{0x22}},
			size:    DEFAULT_SIZE,
			wantErr: true,
		},
		{
			name:    "edit crosses boundary",
			edit:    Edit{9, []byte{0x20, 0x21}, []byte{0x22, 0x23}},
			size:    DEFAULT_SIZE,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.edit.Validate(tt.size)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPatchValidate(t *testing.T) {
	const DEFAULT_PATCH_ID PatchID = "test"

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

	tests := []struct {
		name    string
		patch   Patch
		wantErr bool
	}{
		{
			name: "valid single edit",
			patch: Patch{
				ID:    DEFAULT_PATCH_ID,
				Edits: []Edit{edit(0, 2)},
			},
		},
		{
			name: "valid multiple edits",
			patch: Patch{
				ID:    DEFAULT_PATCH_ID,
				Edits: []Edit{edit(0, 2), edit(5, 2)},
			},
		},
		{
			name: "empty ID",
			patch: Patch{
				Edits: []Edit{edit(0, 2)},
			},
			wantErr: true,
		},
		{
			name: "no edits",
			patch: Patch{
				ID: DEFAULT_PATCH_ID,
			},
			wantErr: true,
		},
		{
			name: "invalid edit",
			patch: Patch{
				ID:    DEFAULT_PATCH_ID,
				Edits: []Edit{edit(9, 2)},
			},
			wantErr: true,
		},
		{
			name: "adjacent edits",
			patch: Patch{
				ID:    DEFAULT_PATCH_ID,
				Edits: []Edit{edit(0, 3), edit(3, 2)},
			},
		},
		{
			name: "overlapping edits",
			patch: Patch{
				ID:    DEFAULT_PATCH_ID,
				Edits: []Edit{edit(0, 3), edit(2, 2)},
			},
			wantErr: true,
		},
		{
			name: "identical ranges",
			patch: Patch{
				ID:    DEFAULT_PATCH_ID,
				Edits: []Edit{edit(0, 3), edit(0, 3)},
			},
			wantErr: true,
		},
		{
			name: "contained edit",
			patch: Patch{
				ID:    DEFAULT_PATCH_ID,
				Edits: []Edit{edit(0, 5), edit(1, 2)},
			},
			wantErr: true,
		},
		{
			name: "unordered overlapping edits",
			patch: Patch{
				ID:    DEFAULT_PATCH_ID,
				Edits: []Edit{edit(2, 3), edit(0, 3)},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.patch.Validate(10)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() = %v, wantErr %v",
					err, tt.wantErr)
			}
		})
	}
}
