# 04 · Soundness & Zero-Knowledge

[← Build a mini-zkVM](03-build-a-mini-zkvm.md) · [Next: The real zkVM landscape →](05-real-zkvms.md)

---

A proof system is only useful if it has two properties:

- **Completeness** — honest proofs always verify.
- **Soundness** — false statements are (almost) always rejected.

And a *zero-knowledge* proof system adds a third:

- **Zero-knowledge** — the proof reveals nothing beyond the truth of the statement.

Let's see which of these our toy achieves, and which it only sketches.

## Completeness ✓

`TestCompleteness` proves an honest Fibonacci execution and verifies it with several query
counts, including a full audit. All accept:

```
--- PASS: TestCompleteness (0.00s)
```

## Soundness — three attacks, three rejections

The test suite plays the cheating prover three different ways. All are caught:

### 1. Tampering with the trace

```go
trace := Execute(program)
trace[5][1].SetUint64(trace[5][1].Uint64() + 1) // corrupt a register mid-execution
proof := proveFromTrace(trace, program, 1, 100000) // full audit
ok, reason := Verify(proof, program, 100000)
// ok == false
```

```
tampered trace correctly rejected: transition constraint violated at step 4
```

Corrupting row 5 breaks the transition *into* it (step 4) and *out of* it (step 5). A full
audit checks every step, so it is caught deterministically. With random sampling it is
caught with probability `1 − (1 − b/n)^q` for `b` bad steps, `n` steps, `q` queries — more
on that limitation below.

### 2. Forging an opened value

```go
op.Values[0]++              // lie about a register value...
proof.Openings[idx] = op    // ...but keep the (now-stale) Merkle path and root
```

```
forged opening correctly rejected: missing or invalid opening for a challenged transition
```

The verifier recomputes the leaf hash from the claimed values and walks the path; a single
changed value no longer hashes to the committed root, so the opening is rejected before any
transition is even checked. This is the **binding** property of the Merkle commitment.

### 3. Lying about the output

```go
proof.ClaimedOutput++       // claim a different result
```

```
wrong output correctly rejected: claimed output does not match the committed final row
```

The claimed output is checked against the committed final row *and* folded into the
Fiat–Shamir seed, so a prover cannot restate the result without invalidating everything.

## Zero-knowledge — sketched, not achieved ⚠️

Be honest with your audience: **this toy is not yet zero-knowledge.** Each opening reveals
the actual register values of the challenged rows. A verifier that queried enough rows could
reconstruct much of the trace.

Real systems close this gap by **randomizing the commitment**: the trace polynomials are
blinded with random values that cancel out of the constraint checks but make the openings
look uniformly random. The verifier still learns "the constraints hold" and nothing else.
Adding blinding here would be a good exercise — the pipeline would not otherwise change.

## The big simplification: random sampling vs. FRI

Our soundness rests on *sampling individual transitions*. To catch a single bad step out of
`n` with high confidence, you need `O(n)` queries — which kills succinctness for large
traces. That is the central problem real STARK-based zkVMs solve with **FRI (Fast
Reed–Solomon Interactive Oracle Proof of Proximity)**:

1. Encode the trace as evaluations of a low-degree polynomial.
2. Express "every transition is correct" as "a certain *constraint polynomial* is divisible
   by the domain's vanishing polynomial" — i.e. it is itself low-degree.
3. A *single* incorrect step makes that polynomial **not** low-degree, and FRI detects
   non-low-degree functions with overwhelming probability from only a **handful** of queries
   — independent of `n`.

So the mental upgrade from this toy to a real STARK zkVM is:

| This toy | Production STARK zkVM |
|----------|------------------------|
| Merkle-commit the raw trace rows | Merkle-commit *polynomial evaluations* of the trace |
| Check sampled transitions directly | Check a constraint polynomial via **FRI** low-degree testing |
| `O(n)` queries for full soundness | `O(λ)` queries (e.g. ~80–100) regardless of `n` |
| Openings reveal values (no ZK) | Blinded commitments give zero-knowledge |

SNARK-based zkVMs make a different choice — they swap Merkle/FRI for a **pairing-based
polynomial commitment** (KZG), trading a trusted setup for tiny constant-size proofs. That
SNARK-vs-STARK fork is the subject of [chapter 06](06-snark-vs-stark.md).

---

[← Build a mini-zkVM](03-build-a-mini-zkvm.md) · [Next: The real zkVM landscape →](05-real-zkvms.md)
