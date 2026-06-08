# 05 · The Real zkVM Landscape

[← Soundness & zero-knowledge](04-soundness-and-zero-knowledge.md) · [Next: SNARK vs STARK in zkVMs →](06-snark-vs-stark.md)

---

Our toy has a 3-instruction ISA and a Merkle-sampling argument. Production zkVMs keep the
*same pipeline* (execute → trace → commit → prove → verify) but scale every part up: real
instruction sets, memory, and FRI/KZG-based low-degree arguments. Here is the field as of
early 2026.

## General-purpose zkVMs

| Project | ISA / language | Proving system | Notes |
|---------|----------------|----------------|-------|
| **RISC Zero** | RISC-V (Rust, C++, Go guests) | STARK, wrappable in Groth16 | Mature; "boundless" proving market; STARK proof can be compressed to a SNARK for cheap on-chain verify |
| **SP1** (Succinct) | RISC-V (Rust) | STARK (Plonky3) + SNARK wrap | Performance-focused, precompiles for hashing/EC ops |
| **Cairo VM** (Starkware) | Cairo assembly (Cairo lang) | STARK | Powers StarkNet; purpose-built ISA designed to be STARK-friendly |
| **Polygon Miden** | Miden assembly | STARK | Stack machine VM, client-side proving focus |
| **Jolt** (a16z) | RISC-V | Lookup-centric (Lasso) + sumcheck | "Just One Lookup Table"; aims for a simpler, faster prover |
| **Valida** (Lita) | Custom RISC-like ISA | STARK (Plonky3) | Designed ground-up for proving efficiency |
| **Nexus** | RISC-V | Folding / recursive | Targets massively parallel proving |
| **OpenVM** (Axiom) | Modular RISC-V + extensions | STARK | Modular: add custom instruction subsystems |
| **zkWASM** (e.g. Delphinus) | WebAssembly | SNARK/STARK | Prove execution of WASM bytecode |

## Libraries & toolkits you'd build *with*

These are the layer beneath a zkVM — the proving-system libraries from the lecture's
"Libraries & Frameworks" slide:

- **Plonky3** — the STARK toolkit (FRI, fields, AIR) underneath SP1 and Valida.
- **gnark** (ConsenSys, **Go**) — Groth16/PLONK SNARKs; the natural Go counterpart to this
  tutorial if you want a *real* SNARK next.
- **arkworks** (Rust) — general-purpose SNARK ecosystem (curves, R1CS, Groth16, Marlin).
- **Halo2** — PLONKish proving used by zcash/Scroll; circuit-centric rather than a VM.
- **Nova / SuperNova** — folding schemes for incrementally verifiable computation.

## How a real proof reaches a blockchain

A subtle but important pattern, and the reason you'll see "STARK **and** SNARK" together:

```
   guest program ─► zkVM executes ─► STARK proof (large, transparent, fast to make)
                                          │
                                          ▼
                              wrap in a SNARK (Groth16/PLONK)
                                          │
                                          ▼
                        tiny constant-size proof ─► verified cheaply on Ethereum
```

STARKs are great to *generate* (no trusted setup, post-quantum, scalable prover) but their
proofs are large. A final SNARK "wrapper" compresses the STARK proof into a few hundred
bytes that an L1 smart contract can verify for minimal gas. This is exactly the "recursive
proving wraps a STARK inside a SNARK" point from the lecture's *Choosing Between SNARK and
STARK* slide.

## Where to go next

- Re-implement the soundness layer with **FRI** instead of row sampling (see
  [chapter 04](04-soundness-and-zero-knowledge.md)) — the single biggest conceptual upgrade.
- Try a real toolchain: **RISC Zero** and **SP1** both have "prove a Rust `fn`" quickstarts
  in their official docs (they require the Rust toolchain, so they are intentionally **not**
  vendored or run in this repo — everything here stays dependency-free and offline-testable).
- For a *Go-native* next step, build a real SNARK circuit with **gnark** and compare its
  trusted-setup + tiny-proof profile against this STARK-style toy.

---

[← Soundness & zero-knowledge](04-soundness-and-zero-knowledge.md) · [Next: SNARK vs STARK in zkVMs →](06-snark-vs-stark.md)
