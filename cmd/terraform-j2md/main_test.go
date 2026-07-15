package main

import (
	"os/exec"
	"testing"
)

func TestVersion(t *testing.T) {
	cmd := exec.Command(
		"go", "run",
		"-ldflags=-X main.Version=0.0.9-io41.1 -X main.Revision=abc1234",
		".", "--version",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run --version: %v\n%s", err, output)
	}
	if got, want := string(output), "0.0.9-io41.1 (abc1234)\n"; got != want {
		t.Errorf("--version = %q, want %q", got, want)
	}
}
