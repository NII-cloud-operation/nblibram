package audit

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func captureStdoutAndStderr(t *testing.T, fn func() error) (string, string, error) {
	t.Helper()
	oldOut := os.Stdout
	oldErr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr
	err := fn()
	wOut.Close()
	wErr.Close()
	os.Stdout = oldOut
	os.Stderr = oldErr
	var bufOut, bufErr bytes.Buffer
	bufOut.ReadFrom(rOut)
	bufErr.ReadFrom(rErr)
	return bufOut.String(), bufErr.String(), err
}

func TestAuditWithLeaks(t *testing.T) {
	// basic.ipynb has a GitHub token in cell 5
	_, _, err := captureStdoutAndStderr(t, func() error {
		return Run([]string{"-file", "../../testdata/basic.ipynb", "-format", "text"})
	})
	if err != ErrLeaksFound {
		t.Errorf("expected ErrLeaksFound, got %v", err)
	}
}

func TestAuditWithLeaksJSON(t *testing.T) {
	out, _, err := captureStdoutAndStderr(t, func() error {
		return Run([]string{"-file", "../../testdata/basic.ipynb", "-format", "json"})
	})
	if err != ErrLeaksFound {
		t.Errorf("expected ErrLeaksFound, got %v", err)
	}
	var findings []map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(out), &findings); jsonErr != nil {
		t.Fatalf("invalid JSON: %v", jsonErr)
	}
	if len(findings) == 0 {
		t.Error("expected at least one finding")
	}
}

func TestAuditCleanNotebook(t *testing.T) {
	clean := `{"cells":[{"cell_type":"code","source":["print(1)"],"outputs":[]}],"nbformat":4,"nbformat_minor":5}`
	old := os.Stdin
	r, w, _ := os.Pipe()
	w.WriteString(clean)
	w.Close()
	os.Stdin = r

	_, stderr, err := captureStdoutAndStderr(t, func() error {
		return Run([]string{"-format", "text"})
	})
	os.Stdin = old

	if err != nil {
		t.Errorf("expected no error for clean notebook, got %v", err)
	}
	if len(stderr) == 0 {
		t.Error("expected 'No leaks detected.' on stderr")
	}
}
