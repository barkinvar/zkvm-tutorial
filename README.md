# Mini-zkVM: A Hands-On Tutorial on Zero-Knowledge Virtual Machines

![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go&logoColor=white)
![Tests](https://img.shields.io/badge/tests-11%20passing-brightgreen)
![Dependencies](https://img.shields.io/badge/dependencies-none-blue)
![License](https://img.shields.io/badge/license-MIT-green)

A small, fully-tested **Zero-Knowledge Virtual Machine** written from scratch in Go,
plus a tutorial series that explains how real zkVMs (RISC Zero, SP1, Cairo, …) work.

The goal is not performance — it is **clarity**. Every moving part of a zkVM is here in
~400 lines of dependency-free Go you can read in one sitting, and a runnable demo that
executes a program, proves it ran correctly, and then catches a cheating prover.

> Companion material for the ZKP lecture, *Tools & Infrastructure → Zero-Knowledge Virtual
> Machines*. See [`docs/06-snark-vs-stark.md`](docs/06-snark-vs-stark.md) for how this maps
> back to the SNARK-vs-STARK trade-offs in the slides.

---

## The big idea

A zkVM lets you run an ordinary program and produce a tiny cryptographic **proof that it
was executed correctly** — which anyone can verify without re-running the program (and,
in a full system, without even seeing the inputs).

```
   program  ─►  ┌───────────┐  execution   ┌──────────┐  proof   ┌──────────┐
                │ VM        │  trace        │  PROVER  │ ───────► │ VERIFIER │ ─► accept / reject
   inputs   ─►  │ (execute) │ ────────────► │ (commit) │          │ (check)  │
                └───────────┘               └──────────┘          └──────────┘
                                          public: program + claimed output
```

This repo implements exactly that pipeline:

1. **VM** — a 4-register machine over a finite field with `SET`, `ADD`, `MUL`.
2. **Execute** — run a program to produce an *execution trace* (the witness).
3. **Commit** — Merkle-hash the trace into a 32-byte root.
4. **Prove** — use Fiat–Shamir to pick random transitions and open them.
5. **Verify** — re-derive the challenges, check the openings against the root, and check
   each opened step obeys the VM's rules.

---

## Quick start

Requires **Go 1.21+** (developed on Go 1.25). No external dependencies.

```bash
# run the whole test suite (VM, Merkle, completeness, soundness)
go test ./... -v

# run the end-to-end demo
go run ./cmd/demo
```

Prefer a one-word runner?

```bash
make demo            # Linux / macOS / CI (needs `make`)
./present.ps1        # Windows PowerShell: vet + tests + demo
```

Presenting this live? See the [presentation cheat sheet](docs/00-cheatsheet.md).

Expected demo output (deterministic):

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

---

## Repository layout

```
zkvm-tutorial/
├── README.md                ← you are here
├── go.mod
├── Makefile                 ← make test / demo / all  (Linux/macOS/CI)
├── present.ps1              ← Windows PowerShell runner
├── minizkvm/                ← the zkVM library (read these in order)
│   ├── vm.go                ← finite field + register VM + execution trace
│   ├── merkle.go            ← Merkle-tree trace commitment
│   ├── proof.go             ← Fiat–Shamir prover & verifier
│   └── zkvm_test.go         ← completeness + soundness tests
├── cmd/demo/main.go         ← runnable end-to-end demo
└── docs/                    ← the tutorial
    ├── 00-cheatsheet.md
    ├── 01-what-is-a-zkvm.md
    ├── 02-the-pipeline.md
    ├── 03-build-a-mini-zkvm.md
    ├── 04-soundness-and-zero-knowledge.md
    ├── 05-real-zkvms.md
    └── 06-snark-vs-stark.md
```

---

## The tutorial

| # | Chapter | What you learn |
|---|---------|----------------|
| 00 | [Presentation cheat sheet](docs/00-cheatsheet.md) | Exact commands + expected output for a live demo |
| 01 | [What is a zkVM?](docs/01-what-is-a-zkvm.md) | The problem zkVMs solve and why "prove a program ran" is powerful |
| 02 | [The proving pipeline](docs/02-the-pipeline.md) | Execute → trace → commit → prove → verify, step by step |
| 03 | [Build a mini-zkVM](docs/03-build-a-mini-zkvm.md) | A guided read of every file in `minizkvm/`, with tested output |
| 04 | [Soundness & zero-knowledge](docs/04-soundness-and-zero-knowledge.md) | Why cheating fails, and what this toy leaves out vs production |
| 05 | [The real zkVM landscape](docs/05-real-zkvms.md) | RISC Zero, SP1, Cairo, Jolt, Valida, gnark, and where to go next |
| 06 | [SNARK vs STARK in zkVMs](docs/06-snark-vs-stark.md) | Mapping this demo back to the lecture's proving-system trade-offs |

---

## What this is — and isn't

**It is** a faithful model of zkVM *architecture*: a real VM, a real trace, a real
cryptographic commitment, and a real Fiat–Shamir non-interactive argument that the trace
follows the VM's transition rules.

**It is not** production-grade. A real zkVM replaces naive random row-sampling with **FRI
low-degree testing** (so a single wrong step is caught with overwhelming probability from a
handful of queries) and adds **zero-knowledge** by randomizing the commitments. Those gaps
are explained honestly in [`docs/04`](docs/04-soundness-and-zero-knowledge.md) — the point
here is to make the skeleton legible before you meet the heavy machinery.

## Make it your own

The module path is a placeholder (`github.com/example/zkvm-tutorial`). To publish it under
your own account, rename the module and update the two internal imports:

```bash
go mod edit -module github.com/<your-username>/zkvm-tutorial
# then update the import in cmd/demo/main.go and minizkvm/example_test.go
go build ./... && go test ./...
```

## License

MIT — see [LICENSE](LICENSE).
