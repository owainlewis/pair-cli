package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadMarkdownInputPreservesStdinBytes(t *testing.T) {
	input := "line one\r\nline two\n\n"
	data, err := ReadMarkdownInput("", "-", strings.NewReader(input))
	if err != nil {
		t.Fatalf("ReadMarkdownInput() error = %v", err)
	}
	if string(data) != input {
		t.Fatalf("expected preserved bytes %q, got %q", input, string(data))
	}
}

func TestReadMarkdownInputBodyAndFileAreMutuallyExclusive(t *testing.T) {
	_, err := ReadMarkdownInput("body", "-", strings.NewReader("stdin"))
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("expected mutual exclusion error, got %v", err)
	}
}

func TestReadMarkdownInputReadsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "doc.md")
	want := []byte("# Title\n")
	if err := os.WriteFile(path, want, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := ReadMarkdownInput("", path, strings.NewReader(""))
	if err != nil {
		t.Fatalf("ReadMarkdownInput() error = %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestConfirmDestructiveRequiresYesWhenNonInteractive(t *testing.T) {
	err := ConfirmDestructive(false, strings.NewReader("yes\n"), &bytes.Buffer{}, "delete?")
	if err == nil || !strings.Contains(err.Error(), "requires --yes") {
		t.Fatalf("expected --yes error, got %v", err)
	}
}

func TestConfirmDestructiveAllowsYesFlag(t *testing.T) {
	if err := ConfirmDestructive(true, strings.NewReader(""), &bytes.Buffer{}, "delete?"); err != nil {
		t.Fatalf("ConfirmDestructive() error = %v", err)
	}
}
