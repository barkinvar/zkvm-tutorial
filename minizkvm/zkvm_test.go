package minizkvm

import (
	"math/big"
	"testing"
)

// The VM executes a simple arithmetic program correctly: (3 + 4) * 5 = 35.
func TestExecuteArithmetic(t *testing.T) {
	program := []Instr{
		{Op: SET, Dst: 0, Imm: 3},     // r0 = 3
		{Op: SET, Dst: 1, Imm: 4},     // r1 = 4
		{Op: ADD, Dst: 2, A: 0, B: 1}, // r2 = 7
		{Op: SET, Dst: 0, Imm: 5},     // r0 = 5
		{Op: MUL, Dst: 3, A: 2, B: 0}, // r3 = 35
	}
	trace := Execute(program)
	got := trace[len(trace)-1][3].Uint64()
	if got != 35 {
		t.Fatalf("expected (3+4)*5 = 35, got %d", got)
	}
}

// The Fibonacci program produces the expected sequence value.
func TestFibonacci(t *testing.T) {
	program := FibonacciProgram(8) // r1 after 8 steps
	trace := Execute(program)
	got := trace[len(trace)-1][1].Uint64()
	// r0=r1=1, then 8 iterations: 1,1->2,3,5,8,13,21,34,55 ; r1 ends at 55.
	if got != 55 {
		t.Fatalf("expected Fibonacci value 55, got %d", got)
	}
}

// Completeness: an honest proof always verifies, both with sampled queries and
// with a full audit.
func TestCompleteness(t *testing.T) {
	program := FibonacciProgram(20)
	for _, nq := range []int{4, 8, 100000 /* full audit */} {
		proof := Prove(program, 1, nq)
		ok, reason := Verify(proof, program, nq)
		if !ok {
			t.Fatalf("honest proof rejected with numQueries=%d: %s", nq, reason)
		}
	}
}

// Soundness 1: a tampered execution trace is rejected. A full audit deterministically
// catches the broken transition.
func TestSoundnessTamperedTrace(t *testing.T) {
	program := FibonacciProgram(10)
	trace := Execute(program)

	// Cheat: corrupt a register in the middle of the trace.
	trace[5][1].SetUint64(trace[5][1].Uint64() + 1)

	proof := proveFromTrace(trace, program, 1, 100000) // full audit
	ok, reason := Verify(proof, program, 100000)
	if ok {
		t.Fatalf("tampered trace was accepted (should be rejected)")
	}
	t.Logf("tampered trace correctly rejected: %s", reason)
}

// Soundness 2: lying about an opened value breaks the Merkle commitment, so the
// verifier rejects even though the root and paths are untouched.
func TestSoundnessForgedOpening(t *testing.T) {
	program := FibonacciProgram(10)
	proof := Prove(program, 1, 100000)

	// Find a non-boundary opening and corrupt one of its values.
	for idx, op := range proof.Openings {
		if idx == 0 || idx == proof.NumRows-1 {
			continue
		}
		op.Values[0]++ // forge a value while keeping the (now-stale) path
		proof.Openings[idx] = op
		break
	}
	ok, reason := Verify(proof, program, 100000)
	if ok {
		t.Fatalf("forged opening was accepted (should be rejected)")
	}
	t.Logf("forged opening correctly rejected: %s", reason)
}

// Soundness 3: claiming the wrong output is rejected because the claim is bound
// into the committed final row (and into the Fiat-Shamir seed).
func TestSoundnessWrongOutput(t *testing.T) {
	program := FibonacciProgram(10)
	proof := Prove(program, 1, 8)
	proof.ClaimedOutput++ // lie about the result
	ok, reason := Verify(proof, program, 8)
	if ok {
		t.Fatalf("wrong claimed output was accepted (should be rejected)")
	}
	t.Logf("wrong output correctly rejected: %s", reason)
}

// Register arithmetic is done in the field F_P, so it wraps around the modulus.
func TestFieldArithmetic(t *testing.T) {
	// (P-1) + 2 == 1 (mod P)
	pMinus1 := new(big.Int).Sub(P, big.NewInt(1))
	if got := fieldAdd(pMinus1, big.NewInt(2)); got.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("field addition did not wrap: got %s", got)
	}
	// (P-1) * (P-1) == 1 (mod P)
	if got := fieldMul(pMinus1, pMinus1); got.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("field multiplication did not wrap: got %s", got)
	}
}

// The commitment is deterministic: the same program yields the same root.
func TestDeterministicCommitment(t *testing.T) {
	program := FibonacciProgram(10)
	a := Prove(program, 1, 6)
	b := Prove(program, 1, 6)
	if string(a.Root) != string(b.Root) {
		t.Fatal("commitment is not deterministic across runs")
	}
}

// Proof size stays small and grows only logarithmically with the trace length.
func TestProofSizeSucceeds(t *testing.T) {
	small := Prove(FibonacciProgram(12), 1, 6).SizeBytes()
	large := Prove(FibonacciProgram(1500), 1, 6).SizeBytes()
	if small <= 0 || large <= 0 {
		t.Fatalf("non-positive proof size: small=%d large=%d", small, large)
	}
	// A 125x larger computation must not produce anywhere near a 125x larger proof.
	if large > small*4 {
		t.Fatalf("proof size grew too fast: small=%d large=%d", small, large)
	}
}

// The Merkle commitment round-trips: a valid opening verifies, a wrong index does not.
func TestMerkleRoundTrip(t *testing.T) {
	leaves := make([][]byte, 6)
	for i := range leaves {
		leaves[i] = leafHash([]byte{byte(i)})
	}
	m := NewMerkle(leaves)
	if !VerifyMerkle(m.Root(), leaves[3], 3, m.Proof(3)) {
		t.Fatal("valid Merkle opening failed to verify")
	}
	if VerifyMerkle(m.Root(), leaves[3], 2, m.Proof(3)) {
		t.Fatal("Merkle opening verified against the wrong index")
	}
}
