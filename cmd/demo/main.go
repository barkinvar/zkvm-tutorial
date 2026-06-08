// Command demo runs the mini-zkVM end to end and prints each stage, so you can
// watch a program get executed, committed, proven, and verified — then watch two
// different cheating provers get caught, and finally see why the proof is
// "succinct" (its size barely grows as the computation gets bigger).
package main

import (
	"encoding/hex"
	"fmt"

	zkvm "github.com/example/zkvm-tutorial/minizkvm"
)

func main() {
	const steps = 12
	const outReg = 1
	const numQueries = 6 // security parameter: how many transitions to challenge

	// 1. A program for our VM: compute a Fibonacci value.
	program := zkvm.FibonacciProgram(steps)
	fmt.Println("== 1. PROGRAM ==")
	fmt.Printf("Fibonacci program: %d instructions, output register r%d\n\n", len(program), outReg)

	// 2. Execute on the VM to get the trace (the witness).
	trace := zkvm.Execute(program)
	result := trace[len(trace)-1][outReg].Uint64()
	fmt.Println("== 2. EXECUTE ==")
	fmt.Printf("Trace has %d rows; computed result = %d\n\n", len(trace), result)

	// 3. Prove: commit to the trace and answer the Fiat-Shamir challenges.
	proof := zkvm.Prove(program, outReg, numQueries)
	fmt.Println("== 3. PROVE ==")
	fmt.Printf("Trace commitment (Merkle root): %s\n", hex.EncodeToString(proof.Root))
	fmt.Printf("Proof opens %d rows for %d challenged transitions (%d bytes)\n\n",
		len(proof.Openings), numQueries, proof.SizeBytes())

	// 4. Verify: re-derive challenges, check openings + transitions.
	ok, reason := zkvm.Verify(proof, program, numQueries)
	fmt.Println("== 4. VERIFY ==")
	fmt.Printf("Honest proof accepted: %v (%s)\n\n", ok, reason)

	// 5. A cheating prover claims a wrong result.
	fmt.Println("== 5. CHEATING: forged output ==")
	badOut := zkvm.Prove(program, outReg, numQueries)
	badOut.ClaimedOutput = result + 1
	ok5, reason5 := zkvm.Verify(badOut, program, numQueries)
	fmt.Printf("Accepted: %v (%s)\n\n", ok5, reason5)

	// 6. A cheating prover lies about an opened trace cell (keeps the stale path).
	fmt.Println("== 6. CHEATING: forged trace cell ==")
	badOpen := zkvm.Prove(program, outReg, numQueries)
	for idx, op := range badOpen.Openings {
		if idx == 0 || idx == badOpen.NumRows-1 {
			continue // skip boundary rows; tamper a challenged transition row
		}
		op.Values[0]++
		badOpen.Openings[idx] = op
		break
	}
	ok6, reason6 := zkvm.Verify(badOpen, program, numQueries)
	fmt.Printf("Accepted: %v (%s)\n\n", ok6, reason6)

	// 7. Succinctness: prove far larger computations with the SAME query budget
	//    and watch the proof size grow only logarithmically (Merkle depth), not
	//    linearly with the number of steps.
	fmt.Println("== 7. SUCCINCTNESS (proof size vs computation size) ==")
	fmt.Printf("%-8s %-12s %-12s\n", "steps", "trace rows", "proof bytes")
	for _, s := range []int{12, 60, 300, 1500} {
		p := zkvm.FibonacciProgram(s)
		pr := zkvm.Prove(p, outReg, numQueries)
		fmt.Printf("%-8d %-12d %-12d\n", s, pr.NumRows, pr.SizeBytes())
	}
	fmt.Println("\nThe computation grows 125x but the proof grows only ~2x: proof size is")
	fmt.Println("logarithmic in the trace length, not linear. That is succinctness.")
}
