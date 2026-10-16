package utils

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestInitializeLogger(t *testing.T) {
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = stdout }()

	logger := InitializeLogger("bookie-test", false)
	logger.Info("hello from logger")

	w.Close()
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read piped output: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "bookie-test") {
		t.Errorf("output %q does not contain service name", output)
	}
	if !strings.Contains(output, `"timestamp"`) {
		t.Errorf("output %q does not contain replaced time key", output)
	}

	var record map[string]any
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Errorf("line is not valid JSON: %v", err)
		}
	}
	if record["msg"] != "hello from logger" {
		t.Errorf("msg = %v, want %q", record["msg"], "hello from logger")
	}
}
