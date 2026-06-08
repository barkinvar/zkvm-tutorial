# 02 · The Proving Pipeline

[← What is a zkVM?](01-what-is-a-zkvm.md) · [Next: Build a mini-zkVM →](03-build-a-mini-zkvm.md)

---

Every zkVM — and our toy — runs the same five-stage pipeline. This chapter explains each
stage conceptually; the next chapter shows the code.

```
 ┌─────────┐   ┌──────────┐   ┌──────────┐   ┌─────────┐   ┌──────────┐
 │ EXECUTE │ ─►│  TRACE   │ ─►│  COMMIT  │ ─►│  PROVE  │ ─►│  VERIFY  │
 └─────────┘   └──────────┘   └──────────┘   └─────────┘   └──────────┘
   run the      table of        Merkle root    open random    re-derive +
   program      every state     of the trace   transitions    check
```

## 1. Execute

Run the program on the VM to produce the **execution trace**: a table with one row per
step, each row being the complete machine state. This is the prover's *witness*. For a
program of `k` instructions the trace has `k + 1` rows (the initial state plus one row after
each instruction).

## 2. Commit

The prover must "lock in" the entire trace **before** it is told what will be checked —
otherwise it could fabricate answers on the fly. We do this with a **Merkle tree**: hash
each row into a leaf, hash pairs of leaves up to a single 32-byte **root**. The root is a
binding commitment: it is infeasible to find a different trace with the same root, yet the
root reveals nothing about the rows. Publishing the root is the prover's promise: *"I have
a specific trace in mind."*

## 3. Challenge (Fiat–Shamir)

In the interactive version of this protocol, the verifier would now flip coins and ask the
prover to reveal a few random transitions. We make it **non-interactive** with the
**Fiat–Shamir transform**: derive the "random" challenges by hashing the commitment plus the
public inputs:

```
seed = H(root ‖ program ‖ claimed_output)
queries = pseudo-random transition indices derived from seed
```

Because `seed` depends on `root`, the prover is bound to its trace before it can know which
rows will be challenged — it cannot cheat adaptively. This is the same trick that turns
interactive SNARKs/STARKs into the one-shot proofs used on blockchains.

## 4. Prove (open)

For each challenged transition `q`, the prover reveals row `q` and row `q+1` together with
their **Merkle authentication paths** (the sibling hashes proving those rows really belong
to the committed root). It also opens the two **boundary** rows: row `0` (to prove execution
started from the agreed state) and the final row (to expose the claimed output). The
collection of openings *is* the proof. Its size depends on the number of queries, **not** on
the length of the trace — that is the succinctness.

## 5. Verify

The verifier never sees the whole trace. It:

1. re-derives the same `seed` and `queries` (it has `root`, `program`, `claimed_output`);
2. checks every opened row against `root` via its Merkle path;
3. checks the **boundary**: row 0 is the all-zero start state, and the final row holds the
   claimed output;
4. checks each **transition**: applying the VM's transition function to the opened row `q`
   reproduces the opened row `q+1`.

If all checks pass, it accepts. A cheating prover fails at step 2 (if it lies about an opened
value, the Merkle path won't match the root) or at step 4 (if the committed trace contains a
bad step that gets challenged).

## Why each guard matters

| Guard | Stops the attack… |
|-------|-------------------|
| Merkle root commits before challenges | …fabricating rows after seeing the questions |
| Fiat–Shamir binds challenges to the root | …grinding for easy challenges |
| Merkle path check on every opening | …revealing values not in the committed trace |
| Boundary check on row 0 | …starting from a convenient fake state |
| Boundary check on final row | …claiming an output the trace doesn't produce |
| Transition check on challenged steps | …a trace that doesn't follow the VM's rules |

The [next chapter](03-build-a-mini-zkvm.md) walks through the Go that implements all six.

---

[← What is a zkVM?](01-what-is-a-zkvm.md) · [Next: Build a mini-zkVM →](03-build-a-mini-zkvm.md)
