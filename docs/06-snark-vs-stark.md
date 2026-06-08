# 06 · SNARK vs STARK in zkVMs

[← The real zkVM landscape](05-real-zkvms.md) · [README →](../README.md)

---

This closing chapter connects the toy back to the lecture's *Tools & Infrastructure →
SNARK vs STARK* slides. The proving system a zkVM chooses is the single biggest factor in
its proof size, cost, and security profile.

## Which family is our toy?

Our mini-zkVM is **STARK-flavoured**:

- It commits with a **hash function** (SHA-256 Merkle tree), not elliptic-curve pairings.
- It therefore needs **no trusted setup** — the only public parameter is "use SHA-256".
- Being hash-based, the approach is **plausibly post-quantum**.
- Its proofs are **larger** than a SNARK's (a set of rows + Merkle paths), and verification
  cost grows with the number of queries.

Swap the Merkle/sampling layer for a **KZG polynomial commitment** and you would have the
core of a **SNARK-based** zkVM instead: tiny constant-size proofs and very cheap
verification, at the price of a one-time **trusted setup** (the "toxic waste"). That is the
exact trade-off table from the slides:

| Property | zk-SNARK | zk-STARK (this toy's family) |
|----------|----------|------------------------------|
| Commitment primitive | EC pairings (KZG) | Hash functions (Merkle + FRI) |
| Trusted setup | Usually required | **None** ← our toy |
| Proof size | ~hundreds of bytes | tens–hundreds of KB |
| Verification cost | Very low (cheap gas) | Higher |
| Post-quantum secure | No | **Yes** ← our toy |
| Prover scaling | Good for small circuits | Quasilinear, large-scale |

## Why production zkVMs often use *both*

As [chapter 05](05-real-zkvms.md) showed, the dominant pattern is **STARK to prove, SNARK to
verify on-chain**:

- Use a STARK to generate the execution proof — no ceremony, scales to millions of trace
  rows, post-quantum.
- **Wrap** that STARK proof in a small SNARK (Groth16/PLONK) so the final artifact is a few
  hundred bytes that an Ethereum contract can verify for minimal gas.

This is precisely the *"the line is blurring — recursive proving wraps a STARK inside a
SNARK"* takeaway from the lecture's *Choosing Between SNARK and STARK* slide.

## The one-paragraph summary for your presentation

> A zkVM proves that a program was executed correctly by committing to its execution trace
> and proving every state transition obeys the VM's rules. STARK-based zkVMs (RISC Zero, SP1,
> Cairo, Miden) commit with hashes — no trusted setup, post-quantum, but larger proofs.
> SNARK-based proving commits with pairings — tiny proofs and cheap verification, but a
> trusted setup. In practice the two are combined: prove with a STARK, then compress with a
> SNARK for on-chain verification. This repo's mini-zkVM is the STARK skeleton, made small
> enough to read end-to-end.

---

[← The real zkVM landscape](05-real-zkvms.md) · [README →](../README.md)
