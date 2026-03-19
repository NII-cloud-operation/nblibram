package notebook

import (
	"encoding/json"
	"os"
	"os/exec"
)

// PklRecord represents a pickled kernel output log entry.
type PklRecord struct {
	MsgType string          `json:"msg_type"`
	Content json.RawMessage `json:"content"`
}

// ReadPklRecord reads a .pkl file and returns the deserialized kernel message.
func ReadPklRecord(path string) (*PklRecord, error) {
	data, err := readPklBytes(path)
	if err != nil {
		return nil, err
	}
	var rec PklRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func readPklBytes(path string) ([]byte, error) {
	cmd := exec.Command("python3", "-c",
		"import pickle,json,sys; obj=pickle.load(open(sys.argv[1],'rb')); json.dump(obj,sys.stdout,default=repr)",
		path)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}
