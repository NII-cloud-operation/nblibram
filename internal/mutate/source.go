package mutate

import (
	"errors"
	"os"
)

func resolveSource(source, sourceFile string) (string, error) {
	if source != "" && sourceFile != "" {
		return "", errors.New("specify --source or --source-file, not both")
	}
	if sourceFile != "" {
		data, err := os.ReadFile(sourceFile)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return source, nil
}
