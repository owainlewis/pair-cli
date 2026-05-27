package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteJSONWritesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, map[string]string{"id": "doc_123"}); err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	var decoded map[string]string
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, buf.String())
	}
	if decoded["id"] != "doc_123" {
		t.Fatalf("unexpected JSON output: %#v", decoded)
	}
}

func TestWriteTableKeepsIDsVisible(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTable(&buf, []string{"ID", "TITLE"}, [][]string{{"task_123", "Draft"}}); err != nil {
		t.Fatalf("WriteTable() error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{"ID", "TITLE", "task_123", "Draft"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in table:\n%s", want, out)
		}
	}
}
