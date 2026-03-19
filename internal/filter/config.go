package filter

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Filters []FilterConfig `toml:"filters"`
}

func LoadConfig(path string) (Config, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Config{}, fmt.Errorf("cannot get home directory: %w", err)
		}
		path = filepath.Join(home, ".nbfilterrc.toml")
	}

	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return Config{}, err
	}
	return config, nil
}

func InitConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot get home directory: %w", err)
	}

	path := filepath.Join(home, ".nbfilterrc.toml")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s", path)
	}

	defaultConfig := `# nbfilter configuration file
# Each filter has a pattern (regex) and a label (# is replaced with a sequential number)

[[filters]]
pattern = '\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b'
label = "[IPv4_#]"

[[filters]]
pattern = '[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*\.(com|org|net|jp|io|dev|local|internal)'
label = "[DOMAIN_#]"
`

	if err := os.WriteFile(path, []byte(defaultConfig), 0644); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Created config file: %s\n", path)
	return nil
}
