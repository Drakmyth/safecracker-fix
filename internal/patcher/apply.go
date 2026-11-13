package patcher

// Apply returns a copy of data with the specified patch fully applied. Partially applied patches are completed.
func (t Target) Apply(data []byte, id PatchID) ([]byte, error) {
	patch, err := t.verifiedPatch(data, id)
	if err != nil {
		return nil, err
	}

	return patch.apply(data), nil
}

// Reverse returns a copy of data with the specified patch fully reversed. Patrially applied patches are also reversed.
func (t Target) Reverse(data []byte, id PatchID) ([]byte, error) {
	patch, err := t.verifiedPatch(data, id)
	if err != nil {
		return nil, err
	}

	return patch.reverse(data), nil
}
