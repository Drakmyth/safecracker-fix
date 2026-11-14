package safecracker

import "testing"

func TestTargetV101Steam(t *testing.T) {
	target := TargetV101Steam()

	if err := target.Validate(); err != nil {
		t.Fatalf("invalid target: %v", err)
	}
}
