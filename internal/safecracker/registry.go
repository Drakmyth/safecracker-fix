package safecracker

import "github.com/Drakmyth/safecracker-fix/internal/patcher"

// SupportedTargets returns all supported executable targets.
func SupportedTargets() []patcher.Target {
	return []patcher.Target{
		TargetV101Steam(),
	}
}
