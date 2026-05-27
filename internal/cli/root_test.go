package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpIncludesCommandGroups(t *testing.T) {
	stdout, stderr, err := executeCommand("--help")
	if err != nil {
		t.Fatalf("expected help to succeed: %v", err)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr, got %q", stderr)
	}

	for _, want := range []string{"auth", "config", "tasks", "docs", "collections"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("expected root help to include %q, got:\n%s", want, stdout)
		}
	}
}

func TestResourceGroupHelpWorks(t *testing.T) {
	for _, args := range [][]string{
		{"tasks", "--help"},
		{"docs", "--help"},
		{"collections", "--help"},
	} {
		stdout, stderr, err := executeCommand(args...)
		if err != nil {
			t.Fatalf("expected %v to succeed: %v", args, err)
		}
		if stderr != "" {
			t.Fatalf("expected no stderr for %v, got %q", args, stderr)
		}
		if !strings.Contains(stdout, "Usage:") {
			t.Fatalf("expected help usage for %v, got:\n%s", args, stdout)
		}
	}
}

func TestGlobalTokenFlagIsNotPrintedByPlaceholder(t *testing.T) {
	const token = "pair_secret_test_token"

	stdout, stderr, err := executeCommand("--token", token, "tasks", "list")
	if err != nil {
		t.Fatalf("expected placeholder command to succeed: %v", err)
	}
	combined := stdout + stderr
	if strings.Contains(combined, token) {
		t.Fatalf("expected token to be redacted from output, got:\n%s", combined)
	}
}

func executeCommand(args ...string) (string, string, error) {
	cmd := NewRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}
