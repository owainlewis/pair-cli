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

func TestGlobalTokenFlagIsNotPrintedByHelp(t *testing.T) {
	const token = "pair_secret_test_token"

	stdout, stderr, err := executeCommand("--token", token, "help")
	if err != nil {
		t.Fatalf("expected placeholder command to succeed: %v", err)
	}
	combined := stdout + stderr
	if strings.Contains(combined, token) {
		t.Fatalf("expected token to be redacted from output, got:\n%s", combined)
	}
}

func TestExecutePrintsCommandErrorsWithoutToken(t *testing.T) {
	const token = "pair_secret_test_token"

	stdout, stderr, code := executeCLI("--token", token, "unknown")
	if code == 0 {
		t.Fatal("expected unknown command to fail")
	}
	if stdout != "" {
		t.Fatalf("expected no stdout, got %q", stdout)
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Fatalf("expected stderr to explain the command error, got %q", stderr)
	}
	if strings.Contains(stderr, token) {
		t.Fatalf("expected token to be redacted from stderr, got %q", stderr)
	}
}

func TestAuthStatusDoesNotPrintToken(t *testing.T) {
	const token = "pair_secret_test_token"
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("PAIR_BASE_URL", "http://localhost:3000")
	t.Setenv("PAIR_TOKEN", token)

	stdout, stderr, code := executeCLI("auth", "status")
	if code != 0 {
		t.Fatalf("expected auth status to succeed, got code %d stderr %q", code, stderr)
	}
	combined := stdout + stderr
	if !strings.Contains(stdout, "base URL: configured") {
		t.Fatalf("expected configured base URL, got %q", stdout)
	}
	if !strings.Contains(stdout, "token: configured") {
		t.Fatalf("expected configured token, got %q", stdout)
	}
	if strings.Contains(combined, token) {
		t.Fatalf("expected token to be redacted from output, got %q", combined)
	}
}

func TestConfigSetTokenDoesNotEchoToken(t *testing.T) {
	const token = "pair_secret_test_token"
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	stdout, stderr, code := executeCLI("config", "set", "token", token)
	if code != 0 {
		t.Fatalf("expected config set token to succeed, got code %d stderr %q", code, stderr)
	}
	combined := stdout + stderr
	if strings.Contains(combined, token) {
		t.Fatalf("expected token to be redacted from output, got %q", combined)
	}
	if !strings.Contains(stdout, "token saved") {
		t.Fatalf("expected save confirmation, got %q", stdout)
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

func executeCLI(args ...string) (string, string, int) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := ExecuteWithInput(args, strings.NewReader(""), &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}
