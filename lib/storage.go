package lib

import (
	"encoding/json"
	"os"
)

type IPData struct {
	IPv4 string `json:"ipv4"`
	IPv6 string `json:"ipv6"`
}

// StoreIPData writes the IPData struct to a JSON file at the given path.
func StoreIPData(path string, data IPData) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// ReadIPData reads the IPData struct from a JSON file at the given path.
func ReadIPData(path string) (IPData, error) {
	var data IPData
	bytes, err := os.ReadFile(path)
	if err != nil {
		return data, err
	}
	err = json.Unmarshal(bytes, &data)
	return data, err
}
