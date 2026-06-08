# present.ps1 - Windows-friendly runner for the mini-zkVM demo.
#
#   ./present.ps1            # vet + full test suite + end-to-end demo
#   ./present.ps1 soundness  # just the soundness tests (the "cheating" story)
#   ./present.ps1 demo       # just the end-to-end demo
#   ./present.ps1 test       # just the full test suite

param([string]$task = "all")

$ErrorActionPreference = "Stop"
Set-Location -Path $PSScriptRoot

function Section($t) { Write-Host "`n=== $t ===" -ForegroundColor Cyan }

switch ($task.ToLower()) {
    "soundness" {
        Section "Soundness tests (cheating provers get caught)"
        go test ./... -run Soundness -v
    }
    "demo" {
        Section "End-to-end demo"
        go run ./cmd/demo
    }
    "test" {
        Section "Full test suite"
        go test ./... -v
    }
    default {
        Section "Static analysis (go vet)"
        go vet ./...
        Section "Full test suite"
        go test ./...
        Section "End-to-end demo"
        go run ./cmd/demo
    }
}
