package helpers

import (
	"encoding/json"
	"io"
	"os"
)

func ReadCfgJSON[T any](path string) (cfg *T, err error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0444) // "./agent.example.cfg.json"
	if err != nil {
		return nil, err
	}

	x, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(x, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
