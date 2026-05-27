package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// ReadMarkdownInput reads markdown from --body, --file, or stdin via --file -.
func ReadMarkdownInput(body, file string, stdin io.Reader) ([]byte, error) {
	if body != "" && file != "" {
		return nil, errors.New("--body and --file are mutually exclusive")
	}
	if body != "" {
		return []byte(body), nil
	}
	if file == "" {
		return nil, errors.New("provide --body or --file")
	}
	if file == "-" {
		data, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return data, nil
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return data, nil
}

// ConfirmDestructive returns nil when a destructive action may proceed.
func ConfirmDestructive(yes bool, stdin io.Reader, stderr io.Writer, prompt string) error {
	if yes {
		return nil
	}

	if file, ok := stdin.(*os.File); !ok || !isTerminal(file) {
		return errors.New("destructive command requires --yes when not running interactively")
	}

	fmt.Fprintf(stderr, "%s [y/N]: ", prompt)
	var answer string
	if _, err := fmt.Fscanln(stdin, &answer); err != nil {
		return fmt.Errorf("read confirmation: %w", err)
	}
	if strings.EqualFold(answer, "y") || strings.EqualFold(answer, "yes") {
		return nil
	}
	return errors.New("aborted")
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
