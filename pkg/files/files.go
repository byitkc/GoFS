package files

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

func CheckSourceFile(path string) error {
	// checkSourceFile will check to see if the file exists
	_, err := os.Stat(path)
	if err == os.ErrNotExist {
		return os.ErrNotExist
	} else if err != nil {
		return err
	}
	return nil
}

func ComputeFileSHA256ToBase64(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to hash file content: %w", err)
	}

	checksum := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(checksum), nil
}
