package tree

import (
	"bytes"
	"fmt"
)

// Tree is a high-level interface to the state tree.
type Tree interface {
	GetEntry(*[20]byte) []byte
	PutEntry(*[20]byte, []byte) bool
	DeleteEntry(*[20]byte) bool
}

// PMTree stores the Patricia-Merkle tree on a key-value storage engine.
// The keys are written in Patricia-order and only one tree revision is kept at a time.
type PMTree struct {
	Store Store
}

// zeroHash is an all zeros Blake2b hash.
// It is used in various places as a placeholder when hashes need recalculation.
var zeroHash = [32]byte{}

func (t *PMTree) GetEntry(key *[20]byte) []byte {
	node := t.Store.GetNode(KeyToNibbles(key))
	switch n := node.(type) {
	case nil:
		return nil
	case *Leaf:
		return n.Value
	default:
		panic("encountered unknown node type")
	}
}

// PutEntry updates the specified key-value pair.
// An empty value marks a deletion.
//
// It returns true when the entry was changed
// and false if it's already in the required state.
func (t *PMTree) PutEntry(key *[20]byte, value []byte) bool {
	changed := t.updateEntry(key, value)
	t.updateHashes(nil)
	return changed
}

func (t *PMTree) updateEntry(key *[20]byte, value []byte) bool {
	prefix := KeyToNibbles(key)
	// Check if the entry already exists.
	node := t.Store.GetNode(prefix)
	if len(value) == 0 && node == nil {
		// The entry was already deleted.
		return false
	} else if leaf, ok := node.(*Leaf); ok {
		// A value at this key already exists.
		if bytes.Equal(leaf.Value, value) {
			// The value that already exists is the same we are trying to insert.
			return false
		}
	}
	t.update(nil, prefix, value, nil)
	return true
}

// Hash returns the root hash.
func (t *PMTree) Hash() [32]byte {
	return t.Store.GetNode(nil).(*Branch).Hash()
}

// update adds or removes a leaf node at the specified node prefix (first argument),
// with the target path at the second argument.
// If value is empty, the leaf node will be deleted, otherwise added.
//
// Recursion is used to deal with branch nodes on the way.
// The recursive function terminates when a new leaf node is added or an existing leaf node is replaced.
// Recursion usually starts when inserting at the root node of the tree.
//
// Whenever it changes nodes, it invalidates hashes in the path by setting them to zero.
// Then, it iterates on all zero nodes and calculates the hashes bottom up.
func (t *PMTree) update(nodePrefix, prefix Nibbles, value []byte, rootPath []Node) {
	if !nodePrefix.PrefixOf(prefix) {
		// Put new account node.
		t.Store.PutNode(prefix, &Leaf{
			Prefix: prefix,
			Value:  value,
		})
		// Put new parent node.
		newParent := &Branch{
			Prefix: nodePrefix.CommonPrefix(prefix),
		}
		newParent.PutChild(nodePrefix[len(newParent.Prefix):], zeroHash)
		newParent.PutChild(prefix[len(newParent.Prefix):], zeroHash)
		t.Store.PutNode(newParent.Prefix, newParent)
		t.invalidatePath(newParent.Prefix, rootPath)
		return
	}
	// Node already exists with given address.
	if bytes.Equal(nodePrefix, prefix) {
		if len(value) == 0 {
			// Delete value and prune branches.
			t.Store.DelNode(nodePrefix)
			t.prune(nodePrefix, rootPath)
		} else {
			// Update value and invalidate hashes.
			t.Store.PutNode(nodePrefix, &Leaf{
				Prefix: nodePrefix,
				Value:  value,
			})
			t.invalidatePath(nodePrefix, rootPath)
		}
		return
	}
	// Node prefix matches, but branch reached. Descend.
	node := t.Store.GetNode(nodePrefix).(*Branch)
	targetChild := node.Children[prefix[len(node.Prefix)]]
	if targetChild.Exists {
		// Child entry already exists where we would descend to.
		// If the existing entry is a branch, we merge it in. Else, replace it.
		rootPath = append(rootPath, node)
		// TODO Instead of copying the path buffer here, do it on the nodes instead.
		childPrefix := append(node.Prefix[:len(node.Prefix):len(node.Prefix)], targetChild.Suffix...)
		t.update(childPrefix, prefix, value, rootPath)
	} else {
		// Child does not exist yet, add it in directly.
		// Write actual child.
		t.Store.PutNode(prefix, &Leaf{
			Prefix: prefix,
			Value:  value,
		})
		// Write updated branch.
		node.PutChild(prefix[len(node.Prefix):], zeroHash)
		t.Store.PutNode(node.Prefix, node)
		// At this point, end recursion, since we reached the end.
		t.invalidatePath(nodePrefix, rootPath)
	}
}

// invalidatePath walks the path leaf to root and invalidates the hashes along the way.
// It also overwrites branch child references as specified in the path.
func (t *PMTree) invalidatePath(prefix Nibbles, path []Node) {
	// Walk the path of replaced branches up to the root.
	current := prefix
	for len(path) > 0 {
		// Take last node in path.
		node := path[len(path)-1].(*Branch)
		path = path[:len(path)-1]
		// Set the child to a zero hash.
		node.PutChild(current[len(node.Prefix):], zeroHash)
		current = node.Prefix
	}
}

// prune walks the path lowest branch to root and
// cleans up any empty branches along the way.
func (t *PMTree) prune(prefix Nibbles, path []Node) {
	current := prefix
	for len(path) > 0 {
		// Take last node in path.
		node := path[len(path)-1].(*Branch)
		path = path[:len(path)-1]
		// Remove item.
		node.Children[current[len(node.Prefix)]].Exists = false
		// Enumerate the number of children.
		children := make([]*Child, 0, 16)
		for i := range node.Children {
			if node.Children[i].Exists {
				children = append(children, &node.Children[i])
			}
		}
		if len(children) == 1 && len(node.Prefix) > 0 {
			// Branches with a single child can be rolled up into the parent.
			//  [Branch 1 (nChildren = X)] => [Branch 2 (nChildren = 1)] => [Node X]
			//    gets transformed to
			//  [Branch 1 (nChildren = X)] => [Node X]
			// Pruning finishes, defer to the non-pruning recursive algorithm to update the path.
			t.Store.DelNode(node.Prefix)
			t.invalidatePath(append(node.Prefix, children[0].Suffix...), path)
			return
		} else if len(children) > 0 || len(node.Prefix) == 0 {
			// Nothing to be deleted, use non-pruning code path.
			t.Store.PutNode(node.Prefix, node)
			t.invalidatePath(node.Prefix, path)
			return
		}
		current = node.Prefix
	}
}

// updateHashes iterates the tree in-order and updates any hashes.
func (t *PMTree) updateHashes(prefix Nibbles) [32]byte {
	node := t.Store.GetNode(prefix)
	switch n := node.(type) {
	case *Leaf:
		return n.Hash()
	case *Branch:
		for i := range n.Children {
			if !n.Children[i].Exists {
				continue
			}
			if n.Children[i].Hash == zeroHash {
				n.Children[i].Hash = t.updateHashes(append(prefix, n.Children[i].Suffix...))
			}
		}
		t.Store.PutNode(n.Prefix, n)
		return n.Hash()
	default:
		panic(fmt.Sprintf("invalid node type in tree: %T", n))
	}
}
