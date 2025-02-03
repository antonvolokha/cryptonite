package steganography

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

const (
	magicBytes = "CRYPTED"
)

func HideInMP3(mp3Path string, data []byte, outputPath string) error {
	// Read MP3 file
	mp3Data, err := os.ReadFile(mp3Path)
	if err != nil {
		return err
	}

	// Create output buffer
	buf := new(bytes.Buffer)
	buf.Write(mp3Data)

	// Write magic bytes
	buf.Write([]byte(magicBytes))

	// Write data size
	binary.Write(buf, binary.LittleEndian, int64(len(data)))

	// Write encrypted data
	buf.Write(data)

	// Save result
	return os.WriteFile(outputPath, buf.Bytes(), 0644)
}

func ExtractFromMP3(path string) ([]byte, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Find magic bytes
	idx := bytes.Index(data, []byte(magicBytes))
	if idx == -1 {
		return nil, errors.New("no hidden data found")
	}

	// Skip magic bytes
	reader := bytes.NewReader(data[idx+len(magicBytes):])

	// Read data size
	var size int64
	if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
		return nil, err
	}

	// Read encrypted data
	result := make([]byte, size)
	if _, err := io.ReadFull(reader, result); err != nil {
		return nil, err
	}

	return result, nil
} 