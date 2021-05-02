package tree

// OverlayStore is an in-memory overlay over an existing Store.
type OverlayStore struct {
	Lower Store
	diffs map[string]Node
}

// GetNode reads a node from the overlay or lower store.
func (o *OverlayStore) GetNode(nbs Nibbles) Node {
	// Node was overridden in overlay.
	override, ok := o.diffs[string(nbs)]
	if ok {
		return override
	}
	// Request from lower layer.
	return o.Lower.GetNode(nbs)
}

// PutNode puts a node in the overlay without affecting the lower store.
func (o *OverlayStore) PutNode(nbs Nibbles, node Node) {
	o.diffs[string(nbs)] = node
}

// DelNode marks a deletion in the overlay without affecting the lower store.
func (o *OverlayStore) DelNode(nbs Nibbles) {
	o.diffs[string(nbs)] = nil
}

// Flush clears and writes down the in-memory overlay to the lower layer.
func (o *OverlayStore) Flush() {
	for key, node := range o.diffs {
		prefix := Nibbles(key)
		if node == nil {
			o.Lower.DelNode(prefix)
		} else {
			o.Lower.PutNode(prefix, node)
		}
		delete(o.diffs, key)
	}
}
