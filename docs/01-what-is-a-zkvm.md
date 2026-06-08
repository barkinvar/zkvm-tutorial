# 01 · What is a Zero-Knowledge Virtual Machine?

[← README](../README.md) · [Next: The proving pipeline →](02-the-pipeline.md)

---

## The problem

Suppose a powerful server runs an expensive computation for you — a rollup executes 10,000
transactions, an ML model scores an input, a game resolves a turn. How do you know it ran
the program **correctly**, without redoing all the work yourself?

Three classic answers, all unsatisfying:

- **Trust the server.** No guarantees.
- **Re-run it yourself.** Defeats the purpose of outsourcing.
- **Have many parties re-run it and vote** (this is what a normal blockchain does). Expensive
  and only as honest as the majority.

A **zkVM** gives a fourth answer: the server returns the result *plus a short proof* that the
program was executed faithfully. Verifying the proof is **exponentially cheaper** than
re-running the program, and the proof reveals nothing beyond "this ran correctly" (optionally
hiding the inputs entirely — the "zero-knowledge" part).

## The key shift: prove a *program*, not a *circuit*

Earlier ZK systems made you express your statement as a hand-built **arithmetic circuit**
(see the lecture's zkDSL section: Circom, Noir, …). That is powerful but painful — every
program must be re-encoded by a cryptography expert.

A zkVM flips the model. You write a program in a normal language (Rust, C, Go, …), compile it
to a small **instruction set** (RISC-V for RISC Zero and SP1, Cairo assembly for StarkNet),
and the zkVM proves the *correct execution of the instruction set itself*. Write any program
→ get a proof, automatically. That generality is why zkVMs are the headline tool in the
"Tools & Infrastructure" part of the course.

```
  zkDSL approach:   your problem ─► hand-written circuit ─► proof
  zkVM  approach:   your program ─► compiled to ISA ─► [VM proven once] ─► proof
```

## What "correct execution" means

Every CPU, real or virtual, works the same way: it holds a **state** (registers, memory, a
program counter) and repeatedly applies a **transition function** — fetch the next
instruction, update the state. Run it for `n` steps and you get a sequence of states:

```
state_0  ──instr_0──►  state_1  ──instr_1──►  state_2  ──►  …  ──►  state_n
```

That sequence is the **execution trace**. "The program ran correctly" is precisely the
statement:

> Every adjacent pair `(state_i, state_{i+1})` is related by the transition function applied
> to `instr_i`, the first state is the agreed starting state, and the last state contains the
> claimed output.

A zkVM is a machine for proving *that* statement succinctly. Everything else — Merkle trees,
polynomials, FRI, pairings — is engineering to make the proof small and the verification fast.

## Our toy VM

To make this concrete, the rest of the tutorial builds a tiny VM:

- **4 registers** `r0..r3`, each holding a finite-field element.
- **3 instructions**: `SET r = imm`, `ADD r = ra + rb`, `MUL r = ra * rb`.
- Execution starts from the **all-zero state**.

It is laughably small, but it has a genuine state, a genuine transition function, and a
genuine execution trace — which is all we need to demonstrate the full proving pipeline.
Real zkVMs differ only in scale: thousands of opcodes, memory, and millions of trace rows.

---

[← README](../README.md) · [Next: The proving pipeline →](02-the-pipeline.md)
