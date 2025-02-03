package steganography

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

const (
	magicBytes  = "CRYPTED"
	id3v2Header = "ID3"
)

// findMP3End finds the actual end of MP3 audio data
func findMP3End(data []byte) int {
	// Look for ID3v1 tag at the end (if exists)
	if len(data) > 128 && bytes.Equal(data[len(data)-128:len(data)-125], []byte("TAG")) {
		return len(data) - 128
	}
	return len(data)
}

// findMP3Start finds the start of actual MP3 audio data
func findMP3Start(data []byte) int {
	// Skip ID3v2 tag if present
	if bytes.HasPrefix(data, []byte(id3v2Header)) {
		// ID3v2 header is 10 bytes, but we need to read the size
		if len(data) < 10 {
			return 0
		}
		// Size is stored in 4 bytes (7 bits each) starting from byte 6
		size := int(data[6])<<21 | int(data[7])<<14 | int(data[8])<<7 | int(data[9])
		return 10 + size
	}
	return 0
}

func HideInMP3(mp3Path string, data []byte, outputPath string) error {
	// Read MP3 file
	mp3Data, err := os.ReadFile(mp3Path)
	if err != nil {
		return err
	}

	// Find actual MP3 data boundaries
	start := findMP3Start(mp3Data)
	end := findMP3End(mp3Data)

	// Create output buffer
	buf := new(bytes.Buffer)

	// Write header if exists
	buf.Write(mp3Data[:start])

	// Write MP3 data
	buf.Write(mp3Data[start:end])

	// Write our hidden data
	buf.Write([]byte(magicBytes))
	binary.Write(buf, binary.LittleEndian, int64(len(data)))
	buf.Write(data)

	// Write ID3v1 tag if exists
	if end < len(mp3Data) {
		buf.Write(mp3Data[end:])
	}

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
