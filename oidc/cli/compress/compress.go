package compress

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// GzipStr gzip-compresses data and returns it as a string.
func GzipStr(data string) (string, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte(data))
	if err != nil {
		return "", fmt.Errorf("gzip write: %w", err)
	}
	err = gz.Close()
	if err != nil {
		return "", fmt.Errorf("gzip close: %w", err)
	}
	return buf.String(), nil
}

// GunzipStr gzip-decompresses data and returns it as a string.
func GunzipStr(data string) (string, error) {
	gzReader, err := gzip.NewReader(bytes.NewReader([]byte(data)))
	if err != nil {
		return "", fmt.Errorf("gzip reader: %w", err)
	}
	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		return "", fmt.Errorf("gzip read: %w", err)
	}
	err = gzReader.Close()
	if err != nil {
		return "", fmt.Errorf("gzip close: %w", err)
	}
	return string(decompressed), nil
}
