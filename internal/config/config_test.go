package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePrecedence(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}
	if err := Save(path, Config{BaseURL: "https://file.example", Token: "pair_file"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	t.Setenv(EnvBaseURL, "https://env.example")
	t.Setenv(EnvToken, "pair_env")

	resolved, err := Resolve(Overrides{BaseURL: "https://flag.example", Token: "pair_flag"})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if resolved.BaseURL != "https://flag.example" || resolved.BaseURLSource != SourceFlag {
		t.Fatalf("expected flag base URL, got %#v", resolved)
	}
	if resolved.Token != "pair_flag" || resolved.TokenSource != SourceFlag {
		t.Fatalf("expected flag token, got %#v", resolved)
	}
}

func TestResolveUsesEnvBeforeFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath() error = %v", err)
	}
	if err := Save(path, Config{BaseURL: "https://file.example", Token: "pair_file"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	t.Setenv(EnvBaseURL, "https://env.example")
	t.Setenv(EnvToken, "pair_env")

	resolved, err := Resolve(Overrides{})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if resolved.BaseURL != "https://env.example" || resolved.BaseURLSource != SourceEnv {
		t.Fatalf("expected env base URL, got %#v", resolved)
	}
	if resolved.Token != "pair_env" || resolved.TokenSource != SourceEnv {
		t.Fatalf("expected env token, got %#v", resolved)
	}
}

func TestSaveWritesOwnerOnlyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pair", "config.json")
	if err := Save(path, Config{Token: "pair_secret"}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("expected mode 0600, got %o", got)
	}
}
