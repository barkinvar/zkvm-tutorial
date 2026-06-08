# 03 · Build a Mini-zkVM

[← The pipeline](02-the-pipeline.md) · [Next: Soundness & zero-knowledge →](04-soundness-and-zero-knowledge.md)

---

This chapter is a guided read of the implementation. Open each file alongside the text.
Everything here is exercised by the test suite — run it first so you trust what follows:

```bash
go test ./... -v
```

```
=== RUN   TestExecuteArithmetic
--- PASS: TestExecuteArithmetic (0.00s)
=== RUN   TestFibonacci
--- PASS: TestFibonacci (0.00s)
=== RUN   TestCompleteness
--- PASS: TestCompleteness (0.00s)
=== RUN   TestSoundnessTamperedTrace
    zkvm_test.go:62: tampered trace correctly rejected: transition constraint violated at step 4
--- PASS: TestSoundnessTamperedTrace (0.00s)
=== RUN   TestSoundnessForgedOpening
    zkvm_test.go:84: forged opening correctly rejected: missing or invalid opening for a challenged transition
--- PASS: TestSoundnessForgedOpening (0.00s)
=== RUN   TestSoundnessWrongOutput
    zkvm_test.go:97: wrong output correctly rejected: claimed output does not match the committed final row
--- PASS: TestSoundnessWrongOutput (0.00s)
=== RUN   TestFieldArithmetic
--- PASS: TestFieldArithmetic (0.00s)
=== RUN   TestDeterministicCommitment
--- PASS: TestDeterministicCommitment (0.00s)
=== RUN   TestProofSizeSucceeds
--- PASS: TestProofSizeSucceeds (0.00s)
=== RUN   TestMerkleRoundTrip
--- PASS: TestMerkleRoundTrip (0.00s)
=== RUN   Example
--- PASS: Example (0.00s)
PASS
ok  	github.com/example/zkvm-tutorial/minizkvm	0.230s
```

---

## `vm.go` — the machine and its trace

The state is four field elements; the field is the **Goldilocks prime** `2^64 − 2^32 + 1`,
the same field used by production STARK zkVMs because arithmetic mod `P` is cheap.

The heart of the file is the **transition function**. It is deliberately a single, pure
function, because it is the *one* rule both the prover and the verifier must agree on:

```go
// Apply runs one instruction on a state and returns the next state.
func Apply(state Row, ins Instr) Row {
	next := state.clone()
	switch ins.Op {
	case SET:
		next[ins.Dst] = new(big.Int).Mod(new(big.Int).SetUint64(ins.Imm), P)
	case ADD:
		next[ins.Dst] = fieldAdd(state[ins.A], state[ins.B])
	case MUL:
		next[ins.Dst] = fieldMul(state[ins.A], state[ins.B])
	}
	return next
}
```

`Execute` just folds `Apply` over the program, starting from the all-zero state, recording
every intermediate state:

```go
func Execute(program []Instr) []Row {
	trace := []Row{zeroRow()}
	for _, ins := range program {
		trace = append(trace, Apply(trace[len(trace)-1], ins))
	}
	return trace
}
```

That slice of rows is the execution trace — the witness we are going to commit to.

## `merkle.go` — committing to the trace

Each row is serialized to fixed-width bytes and hashed into a leaf; pairs of nodes are
hashed up to a single root. Domain-separation tags (`0x00` for leaves, `0x01` for internal
nodes) prevent a leaf from being reinterpreted as an internal node.

```go
func leafHash(data []byte) []byte { h := sha256.Sum256(append([]byte{0x00}, data...)); return h[:] }
func nodeHash(l, r []byte) []byte { /* sha256(0x01 ‖ l ‖ r) */ }
```

`Proof(index)` returns the sibling hashes from leaf to root, and `VerifyMerkle` recomputes
the root from a leaf + path. The `TestMerkleRoundTrip` test checks that a valid opening
verifies and that the *same* opening fails against the wrong index — i.e. the position is
authenticated, not just the value.

## `proof.go` — prover and verifier

**Fiat–Shamir** turns the commitment + public inputs into challenges:

```go
func fiatShamirSeed(root []byte, program []Instr, claimed uint64, outReg int) []byte {
	h := sha256.New()
	h.Write(root)
	h.Write(progBytes(program))
	// ... also bind the claimed output and output register ...
	return h.Sum(nil)
}
```

`deriveQueries` expands that seed into transition indices. A nice property for teaching: if
you ask for at least as many queries as there are transitions, it simply **audits every
step** (deterministic and fully sound) — handy for the soundness tests.

`Prove` ties it together: execute, Merkle-commit, derive queries, and open the challenged
transitions plus the two boundary rows.

`Verify` re-derives the challenges and runs the six guards from
[chapter 02](02-the-pipeline.md). The transition guard is the punchline — it reuses the
*exact same* `Apply` the VM used:

```go
for _, q := range queries {
	before, ok1 := open(q)
	after,  ok2 := open(q + 1)
	if !ok1 || !ok2 { return false, "missing or invalid opening for a challenged transition" }
	expected := Apply(before, program[q])
	for i := range expected {
		if expected[i].Cmp(after[i]) != 0 {
			return false, "transition constraint violated at step " + strconv.Itoa(q)
		}
	}
}
```

## `cmd/demo/main.go` — see it run

```bash
go run ./cmd/demo
```

```
== 1. PROGRAM ==
Fibonacci program: 38 instructions, output register r1

== 2. EXECUTE ==
Trace has 39 rows; computed result = 377

== 3. PROVE ==
Trace commitment (Merkle root): 729fa31f4a3a217e53ea3fe27072361bd6786a7795f35c86b0831cf8ef1f6ec0
Proof opens 14 rows for 6 challenged transitions (3280 bytes)

== 4. VERIFY ==
Honest proof accepted: true (ok)

== 5. CHEATING: forged output ==
Accepted: false (claimed output does not match the committed final row)

== 6. CHEATING: forged trace cell ==
Accepted: false (missing or invalid opening for a challenged transition)

== 7. SUCCINCTNESS (proof size vs computation size) ==
steps    trace rows   proof bytes
12       39           3280
60       183          4176
300      903          5072
1500     4503         6416

The computation grows 125x but the proof grows only ~2x: proof size is
logarithmic in the trace length, not linear. That is succinctness.
```

Section 7 makes the defining property concrete: as the computation grows **125x** (12 → 1500
steps, 39 → 4503 trace rows) the proof grows only from 3280 to 6416 bytes — about **2x**.
That sub-linear (logarithmic) growth comes from the Merkle path length; the verifier's work
is governed by the query count, not the size of the computation.

The [next chapter](04-soundness-and-zero-knowledge.md) digs into *why the cheating attempt
fails*, and is honest about what this toy still leaves out.

---

[← The pipeline](02-the-pipeline.md) · [Next: Soundness & zero-knowledge →](04-soundness-and-zero-knowledge.md)
