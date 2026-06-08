package minizkvm

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"strconv"
)

// Opening reveals one trace row and its Merkle authentication path.
type Opening struct {
	Index  int
	Values []uint64
	Path   [][]byte
}

// Proof is what the prover sends the verifier. It is succinct: its size grows
// with the number of queries (a security parameter), not with the trace length.
type Proof struct {
	Root          []byte
	NumRows       int
	OutReg        int
	ClaimedOutput uint64
	Openings      map[int]Opening
}

// SizeBytes returns the wire size of the proof: the commitment plus every
// opening (its revealed register values and its Merkle path). This grows with
// the number of queries and the tree depth (logarithmic in the trace length),
// not with the length of the computation itself — which is what makes the proof
// "succinct".
func (p Proof) SizeBytes() int {
	n := len(p.Root)
	for _, op := range p.Openings {
		n += len(op.Values) * 8 // register values
		for _, h := range op.Path {
			n += len(h) // sibling hashes
		}
		n += 8 // leaf index
	}
	return n
}

func progBytes(program []Instr) []byte {
	var b []byte
	for _, ins := range program {
		b = append(b, ins.Op...)
		x := make([]byte, 32)
		binary.BigEndian.PutUint64(x[0:], uint64(ins.Dst))
		binary.BigEndian.PutUint64(x[8:], uint64(ins.A))
		binary.BigEndian.PutUint64(x[16:], uint64(ins.B))
		binary.BigEndian.PutUint64(x[24:], ins.Imm)
		b = append(b, x...)
	}
	return b
}

// fiatShamirSeed binds the random challenges to the commitment and public
// inputs, turning the interactive protocol into a non-interactive one. Because
// the seed depends on the root, the prover must commit before it learns which
// rows will be challenged.
func fiatShamirSeed(root []byte, program []Instr, claimed uint64, outReg int) []byte {
	h := sha256.New()
	h.Write(root)
	h.Write(progBytes(program))
	tmp := make([]byte, 16)
	binary.BigEndian.PutUint64(tmp[0:], claimed)
	binary.BigEndian.PutUint64(tmp[8:], uint64(outReg))
	h.Write(tmp)
	return h.Sum(nil)
}

// deriveQueries selects which transitions to challenge. If numQueries reaches
// the number of transitions it audits all of them (deterministic, fully sound);
// otherwise it derives distinct pseudo-random indices from the seed.
func deriveQueries(seed []byte, numTransitions, numQueries int) []int {
	if numQueries >= numTransitions {
		all := make([]int, numTransitions)
		for i := range all {
			all[i] = i
		}
		return all
	}
	seen := map[int]bool{}
	var qs []int
	for ctr := uint32(0); len(qs) < numQueries; ctr++ {
		h := sha256.New()
		h.Write(seed)
		c := make([]byte, 4)
		binary.BigEndian.PutUint32(c, ctr)
		h.Write(c)
		d := h.Sum(nil)
		idx := int(binary.BigEndian.Uint64(d[:8]) % uint64(numTransitions))
		if !seen[idx] {
			seen[idx] = true
			qs = append(qs, idx)
		}
	}
	return qs
}

func rowToUint64(r Row) []uint64 {
	out := make([]uint64, len(r))
	for i, v := range r {
		out[i] = v.Uint64()
	}
	return out
}

func uint64ToRow(u []uint64) Row {
	r := make(Row, len(u))
	for i, v := range u {
		r[i] = new(big.Int).SetUint64(v)
	}
	return r
}

// Prove executes the program and produces a proof of correct execution.
func Prove(program []Instr, outReg, numQueries int) Proof {
	return proveFromTrace(Execute(program), program, outReg, numQueries)
}

// proveFromTrace lets tests inject a (possibly tampered) trace to exercise
// soundness. Honest provers always go through Prove.
func proveFromTrace(trace []Row, program []Instr, outReg, numQueries int) Proof {
	leaves := make([][]byte, len(trace))
	for i, row := range trace {
		leaves[i] = leafHash(rowBytes(row))
	}
	m := NewMerkle(leaves)
	root := m.Root()
	last := len(trace) - 1
	claimed := trace[last][outReg].Uint64()

	seed := fiatShamirSeed(root, program, claimed, outReg)
	queries := deriveQueries(seed, len(program), numQueries)

	need := map[int]bool{0: true, last: true}
	for _, q := range queries {
		need[q] = true
		need[q+1] = true
	}
	openings := map[int]Opening{}
	for idx := range need {
		openings[idx] = Opening{Index: idx, Values: rowToUint64(trace[idx]), Path: m.Proof(idx)}
	}
	return Proof{Root: root, NumRows: len(trace), OutReg: outReg, ClaimedOutput: claimed, Openings: openings}
}

// Verify re-derives the Fiat-Shamir challenges and checks (1) every opened row
// is consistent with the committed root, (2) the boundary conditions, and
// (3) each challenged transition obeys the VM's transition function.
// It returns a boolean and a human-readable reason.
func Verify(proof Proof, program []Instr, numQueries int) (bool, string) {
	numTransitions := len(program)
	if proof.NumRows != numTransitions+1 {
		return false, "row count does not match program length"
	}
	seed := fiatShamirSeed(proof.Root, program, proof.ClaimedOutput, proof.OutReg)
	queries := deriveQueries(seed, numTransitions, numQueries)

	open := func(idx int) (Row, bool) {
		op, ok := proof.Openings[idx]
		if !ok {
			return nil, false
		}
		row := uint64ToRow(op.Values)
		if !VerifyMerkle(proof.Root, leafHash(rowBytes(row)), idx, op.Path) {
			return nil, false
		}
		return row, true
	}

	// Boundary: execution must start from the all-zero state.
	row0, ok := open(0)
	if !ok {
		return false, "missing or invalid opening for the initial row"
	}
	for _, v := range row0 {
		if v.Sign() != 0 {
			return false, "initial state is not all-zero"
		}
	}

	// Boundary: the committed final row must match the claimed output.
	last := proof.NumRows - 1
	rowLast, ok := open(last)
	if !ok {
		return false, "missing or invalid opening for the final row"
	}
	if rowLast[proof.OutReg].Uint64() != proof.ClaimedOutput {
		return false, "claimed output does not match the committed final row"
	}

	// Transition constraints at every challenged step.
	for _, q := range queries {
		before, ok1 := open(q)
		after, ok2 := open(q + 1)
		if !ok1 || !ok2 {
			return false, "missing or invalid opening for a challenged transition"
		}
		expected := Apply(before, program[q])
		for i := range expected {
			if expected[i].Cmp(after[i]) != 0 {
				return false, "transition constraint violated at step " + strconv.Itoa(q)
			}
		}
	}
	return true, "ok"
}
