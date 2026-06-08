// Package minizkvm is a teaching-sized Zero-Knowledge Virtual Machine.
//
// It models the architecture every real zkVM (RISC Zero, SP1, Cairo, ...) shares:
//
//	program --> [VM executes] --> execution trace --> [commit + prove] --> proof
//	                                                                        |
//	                                       public program + output <--- [verify]
//
// The VM is a tiny register machine over a finite field; the prover commits to
// the full execution trace with a Merkle tree and answers Fiat-Shamir challenges
// by opening individual trace rows; the verifier re-derives the challenges and
// checks that each opened transition obeys the VM's rules.
package minizkvm

import (
	"encoding/binary"
	"math/big"
)

// P is the Goldilocks prime 2^64 - 2^32 + 1, a field used by real STARK-based
// zkVMs because arithmetic mod P is fast. All register values live in F_P.
var P = new(big.Int).SetUint64(18446744069414584321)

func fieldAdd(a, b *big.Int) *big.Int { return new(big.Int).Mod(new(big.Int).Add(a, b), P) }
func fieldMul(a, b *big.Int) *big.Int { return new(big.Int).Mod(new(big.Int).Mul(a, b), P) }

// Opcodes supported by the VM.
const (
	SET = "SET" // reg[Dst] = Imm
	ADD = "ADD" // reg[Dst] = reg[A] + reg[B]
	MUL = "MUL" // reg[Dst] = reg[A] * reg[B]
)

// Instr is a single VM instruction.
type Instr struct {
	Op   string
	Dst  int
	A, B int
	Imm  uint64
}

// NumRegs is the register-file size of our tiny VM.
const NumRegs = 4

// Row is the full register state at one step of execution.
type Row []*big.Int

func zeroRow() Row {
	r := make(Row, NumRegs)
	for i := range r {
		r[i] = big.NewInt(0)
	}
	return r
}

func (r Row) clone() Row {
	c := make(Row, len(r))
	for i := range r {
		c[i] = new(big.Int).Set(r[i])
	}
	return c
}

// Apply runs one instruction on a state and returns the next state. This pure
// transition function is the single source of truth shared by prover and
// verifier: proving correct execution means proving every consecutive pair of
// trace rows is related by Apply.
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

// Execute runs a program from the all-zero state and returns the full trace:
// len(program)+1 rows (the initial row plus one row after each instruction).
func Execute(program []Instr) []Row {
	trace := []Row{zeroRow()}
	for _, ins := range program {
		trace = append(trace, Apply(trace[len(trace)-1], ins))
	}
	return trace
}

// rowBytes serializes a row to fixed-width big-endian bytes for hashing.
func rowBytes(r Row) []byte {
	buf := make([]byte, 8*len(r))
	for i, v := range r {
		binary.BigEndian.PutUint64(buf[i*8:], v.Uint64())
	}
	return buf
}

// FibonacciProgram builds a program that computes Fibonacci numbers using
// r0=prev, r1=cur, r2=temp, r3=zero (never written, used to copy registers).
// After `steps` iterations r1 holds the running Fibonacci value.
func FibonacciProgram(steps int) []Instr {
	prog := []Instr{
		{Op: SET, Dst: 0, Imm: 1}, // r0 = 1
		{Op: SET, Dst: 1, Imm: 1}, // r1 = 1
	}
	for i := 0; i < steps; i++ {
		prog = append(prog,
			Instr{Op: ADD, Dst: 2, A: 0, B: 1}, // r2 = r0 + r1  (next)
			Instr{Op: ADD, Dst: 0, A: 1, B: 3}, // r0 = r1 + 0   (prev <- cur)
			Instr{Op: ADD, Dst: 1, A: 2, B: 3}, // r1 = r2 + 0   (cur  <- next)
		)
	}
	return prog
}
