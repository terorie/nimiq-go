package tree

// MemStore is a store backed by a Go map in memory.
type MemStore struct {
	nodes map[string]Node
}

// NewMemStore creates a new, empty MemStore.
func NewMemStore() *MemStore {
	return &MemStore{
		nodes: map[string]Node{
			"": &Branch{},
		},
	}
}

// GetNode returns a node previously inserted with PutNode.
func (m *MemStore) GetNode(nbs Nibbles) Node {
	return m.nodes[string(nbs)]
	//return m.nodes[nbs.String()]
}

// PutNode sets a node under the path described by Nibbles.
func (m *MemStore) PutNode(nbs Nibbles, node Node) {
	m.nodes[string(nbs)] = node
	//m.nodes[nbs.String()] = node
}

// DelNode removes an existing node or does nothing when the node did not exist.
func (m *MemStore) DelNode(nbs Nibbles) {
	delete(m.nodes, string(nbs))
	//delete(m.nodes, nbs.String())
}
