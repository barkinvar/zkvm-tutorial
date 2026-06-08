# 00 · Presentation Cheat Sheet

Exact commands to run live during the talk. Run them from the repo root
(`zkvm-tutorial/`). Requires **Go 1.21+** — nothing else.

> Tip: `go run`/`go test` print nothing until they finish compiling (a second or two the
> first time). Run each once before the talk so Go's build cache is warm and output is instant.

---

## 0. One-line sanity check (do this before you present)

```bash
go test ./...
```

Expected:

```
ok  	github.com/example/zkvm-tutorial/minizkvm	0.2s
```

## 1. Show the VM proving and verifying a program

```bash
go run ./cmd/demo
```

Expected output:

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

**What to say:** "The VM runs a Fibonacci program and commits to the whole execution as a
single 32-byte Merkle root. The honest proof verifies; two different cheats are rejected;
and section 7 is the punchline — a 125x bigger computation makes a proof only ~2x bigger."

## 2. Show the cheating provers getting caught (the soundness story)

```bash
go test ./... -run Soundness -v
```

Expected:

```
=== RUN   TestSoundnessTamperedTrace
    zkvm_test.go:62: tampered trace correctly rejected: transition constraint violated at step 4
--- PASS: TestSoundnessTamperedTrace (0.00s)
=== RUN   TestSoundnessForgedOpening
    zkvm_test.go:84: forged opening correctly rejected: missing or invalid opening for a challenged transition
--- PASS: TestSoundnessForgedOpening (0.00s)
=== RUN   TestSoundnessWrongOutput
    zkvm_test.go:97: wrong output correctly rejected: claimed output does not match the committed final row
--- PASS: TestSoundnessWrongOutput (0.00s)
PASS
```

**What to say:** "Three different ways to cheat — corrupt the trace, forge an opened value,
or lie about the output — and all three are rejected."

## 3. Full test suite (if asked "is it all tested?")

```bash
go test ./... -v
```

Eleven checks: VM arithmetic, Fibonacci, completeness, three soundness attacks, field
wraparound, deterministic commitment, proof size, Merkle round-trip, and a runnable
doc `Example` — all `PASS`.

---

## Shortcut runners

If you'd rather not type the Go commands:

```bash
# Linux / macOS / CI (needs `make`)
make demo        # run the end-to-end demo
make test        # run the suite
make             # fmt + vet + test + demo

# Windows PowerShell (no make needed)
./present.ps1            # vet + test + demo, with section headers
./present.ps1 soundness  # just the soundness tests
```

---

## 60-second script

1. *"A zkVM proves a program ran correctly without re-running it."* → `go run ./cmd/demo`
2. Point at the **Merkle root**: *"the entire execution squashed into 32 bytes."*
3. Point at **accepted: true** then **forged: false**.
4. *"And here's why you can't cheat:"* → `go test ./... -run Soundness -v`
5. Tie back to the slides: *"this is the STARK skeleton — hash commitments, no trusted
   setup; swap in KZG pairings and you get a SNARK."* (see
   [06-snark-vs-stark.md](06-snark-vs-stark.md))
