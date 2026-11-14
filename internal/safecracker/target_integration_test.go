//go:build integration

package safecracker

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"github.com/Drakmyth/safecracker-fix/internal/patcher"
)

func TestTargetV101SteamIntegration(t *testing.T) {
	const (
		FIXTURE_PATH          = "testdata/v101_steam/Safecracker.exe"
		EXPECTED_PATCHED_HASH = "3ce6c4c4525ddc9e500d7fa7fa03f900f56dcd0398f43bba938226c866a2d1a3"
	)

	original, err := os.ReadFile(FIXTURE_PATH)
	if err != nil {
		t.Fatalf("read integration fixture: %v", err)
	}

	target := TargetV101Steam()

	// Verify the original executable.
	inspection, err := target.Inspect(original)
	if err != nil {
		t.Fatalf("inspect original: %v", err)
	}

	if got := inspection.Patches[PATCH_FPU_PRESERVE]; got != patcher.PATCH_STATE_UNAPPLIED {
		t.Fatalf("original patch state = %v, want Unapplied", got)
	}

	// Apply the performance patch.
	patched, err := target.Apply(original, PATCH_FPU_PRESERVE)
	if err != nil {
		t.Fatalf("apply patch: %v", err)
	}

	// Verify the exact patched binary.
	hash := fmt.Sprintf("%x", sha256.Sum256(patched))
	if hash != EXPECTED_PATCHED_HASH {
		t.Fatalf("patched hash = %s, want %s",
			hash, EXPECTED_PATCHED_HASH)
	}

	// Verify the patched executable is recognized.
	inspection, err = target.Inspect(patched)
	if err != nil {
		t.Fatalf("inspect patched: %v", err)
	}

	if got := inspection.Patches[PATCH_FPU_PRESERVE]; got != patcher.PATCH_STATE_APPLIED {
		t.Fatalf("patched state = %v, want Applied", got)
	}

	// Reverse the patch.
	restored, err := target.Reverse(patched, PATCH_FPU_PRESERVE)
	if err != nil {
		t.Fatalf("reverse patch: %v", err)
	}

	// Verify exact restoration.
	if !bytes.Equal(restored, original) {
		t.Fatal("restored executable differs from original")
	}
}
