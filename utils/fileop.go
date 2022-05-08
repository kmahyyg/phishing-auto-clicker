package utils

import (
	"os"
	"path/filepath"
)

// CheckExists checks if a file or directory exists
func CheckExists(path string) (bool, int) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	fdinfo, err := os.Stat(abspath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, -1
		}
		return false, -1
	}
	if fdinfo.IsDir() {
		return true, 1
	}
	return true, 0
}

// XORStream uses XOR to transform data in a simple but not secure way
func XORStream(key []byte, data []byte) []byte {
	for i := 0; i < len(data); i++ {
		data[i] = data[i] ^ key[0]
	}
	return data
}
