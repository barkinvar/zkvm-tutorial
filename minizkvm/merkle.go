package minizkvm

import "crypto/sha256"

// The trace commitment is a Merkle tree over the trace rows. The root is a
// short (32-byte) binding commitment to the entire execution; an opening is a
// row plus the sibling hashes needed to recompute the root.

func leafHash(data []byte) []byte {
	h := sha256.Sum256(append([]byte{0x00}, data...)) // 0x00 = leaf domain tag
	return h[:]
}

func nodeHash(l, r []byte) []byte {
	buf := make([]byte, 0, 1+len(l)+len(r))
	buf = append(buf, 0x01) // 0x01 = internal-node domain tag
	buf = append(buf, l...)
	buf = append(buf, r...)
	h := sha256.Sum256(buf)
	return h[:]
}

// Merkle is a fixed binary Merkle tree (leaves padded to a power of two).
type Merkle struct {
	levels [][][]byte // levels[0] = leaves
}

func nextPow2(n int) int {
	p := 1
	for p < n {
		p <<= 1
	}
	return p
}

// NewMerkle builds a tree over the given leaf hashes.
func NewMerkle(leaves [][]byte) *Merkle {
	size := nextPow2(len(leaves))
	padded := make([][]byte, size)
	empty := leafHash(nil)
	for i := 0; i < size; i++ {
		if i < len(leaves) {
			padded[i] = leaves[i]
		} else {
			padded[i] = empty
		}
	}
	levels := [][][]byte{padded}
	for len(levels[len(levels)-1]) > 1 {
		cur := levels[len(levels)-1]
		next := make([][]byte, len(cur)/2)
		for i := 0; i < len(cur); i += 2 {
			next[i/2] = nodeHash(cur[i], cur[i+1])
		}
		levels = append(levels, next)
	}
	return &Merkle{levels: levels}
}

// Root returns the commitment.
func (m *Merkle) Root() []byte { return m.levels[len(m.levels)-1][0] }

// Proof returns the sibling path (bottom-up) authenticating a leaf index.
func (m *Merkle) Proof(index int) [][]byte {
	var path [][]byte
	idx := index
	for lvl := 0; lvl < len(m.levels)-1; lvl++ {
		path = append(path, m.levels[lvl][idx^1])
		idx /= 2
	}
	return path
}

// VerifyMerkle recomputes the root from a leaf hash and its sibling path and
// reports whether it matches the committed root.
func VerifyMerkle(root, leaf []byte, index int, path [][]byte) bool {
	h := leaf
	idx := index
	for _, sib := range path {
		if idx&1 == 0 {
			h = nodeHash(h, sib)
		} else {
			h = nodeHash(sib, h)
		}
		idx /= 2
	}
	return string(h) == string(root)
}
