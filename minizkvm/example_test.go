package minizkvm_test

import (
	"fmt"

	zkvm "github.com/example/zkvm-tutorial/minizkvm"
)

// Example shows the whole public API: build a program, prove it ran, and verify
// the proof. (This example is executed by `go test`, so its output is checked.)
func Example() {
	// A program for the VM: compute a Fibonacci value, leaving the result in r1.
	program := zkvm.FibonacciProgram(8)

	// Prove correct execution, challenging 4 random transitions.
	const outReg, numQueries = 1, 4
	proof := zkvm.Prove(program, outReg, numQueries)

	// Verify re-derives the challenges and checks the proof.
	ok, reason := zkvm.Verify(proof, program, numQueries)

	fmt.Printf("result=%d verified=%v (%s)\n", proof.ClaimedOutput, ok, reason)
	// Output:
	// result=55 verified=true (ok)
}
