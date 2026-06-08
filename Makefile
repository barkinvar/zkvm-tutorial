# Mini-zkVM tutorial — convenience targets (Linux / macOS / CI).
# Windows users without `make`: use ./present.ps1 instead.

.PHONY: all test demo vet fmt clean

all: fmt vet test demo

test:        ## run the full test suite
	go test ./...

test-v:      ## run the full test suite, verbose
	go test ./... -v

soundness:   ## run only the soundness tests, verbose
	go test ./... -run Soundness -v

demo:        ## run the end-to-end demo
	go run ./cmd/demo

vet:         ## static analysis
	go vet ./...

fmt:         ## format all Go sources
	gofmt -w .

clean:       ## remove build artifacts
	go clean ./...
