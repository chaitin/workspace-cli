package chaitin

import (
	"bytes"
	"testing"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs(nil)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() returned error: %v", err)
	}

	got := stdout.String()
	want := "Uncomputable, infinite possibilities\n"
	if got != want {
		t.Fatalf("unexpected output: got %q, want %q", got, want)
	}
}
