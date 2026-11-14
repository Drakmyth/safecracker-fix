package safecracker

import "github.com/Drakmyth/safecracker-fix/internal/patcher"

const (
	TARGET_V101_STEAM  patcher.TargetID = "v1.01-steam"
	PATCH_FPU_PRESERVE patcher.PatchID  = "fpu-preserve"
)

func TargetV101Steam() patcher.Target {
	return patcher.Target{
		ID:   TARGET_V101_STEAM,
		Name: "Safecracker v1.01 (Steam)",
		Size: 4788224,
		OriginalHash: [32]byte{
			0x66, 0x62, 0xFB, 0xB3, 0x06, 0xD2, 0x5D, 0x2E,
			0x02, 0x6C, 0xD8, 0x7C, 0xBC, 0xF8, 0xC6, 0xD2,
			0x9A, 0x1E, 0x3C, 0x9D, 0x42, 0xB7, 0x35, 0x60,
			0xE2, 0x69, 0x46, 0xDB, 0xAC, 0xCF, 0x6F, 0x7E,
		},
		Patches: []patcher.Patch{
			{
				ID:   PATCH_FPU_PRESERVE,
				Name: "High Uptime Performance Fix",
				Edits: []patcher.Edit{
					{
						Offset:      0x38FC3,
						Original:    []byte{0x83, 0xC5, 0x20},
						Replacement: []byte{0x83, 0xC5, 0x22},
					},
				},
			},
		},
	}
}
