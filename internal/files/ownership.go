package files

import (
	"encoding/json"
	"github.com/ntnyq/rivu/internal/storage"
	"os"
)

func ReadOwnership(path string) (map[string]string, error) {
	out := map[string]string{}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return out, nil
	}
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &out)
	return out, err
}
func SaveOwnership(path string, owners map[string]string) error {
	data, err := json.Marshal(owners)
	if err != nil {
		return err
	}
	return storage.WriteFile(path, data, 0o600)
}
