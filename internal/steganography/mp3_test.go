package steganography

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestMP3Steganography(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()

	// Create a minimal valid MP3 file with ID3v2 and ID3v1 tags
	mp3Data := []byte{
		// ID3v2 header
		0x49, 0x44, 0x33, // "ID3"
		0x03, 0x00, // version
		0x00,                   // flags
		0x00, 0x00, 0x00, 0x00, // size
		// MP3 frame header
		0xFF, 0xFB, 0x90, 0x64, // MPEG1 Layer3
		// Some audio data
		0x01, 0x02, 0x03, 0x04,
		// ID3v1 tag
		0x54, 0x41, 0x47, // "TAG"
	}

	mp3Path := filepath.Join(tmpDir, "test.mp3")
	if err := os.WriteFile(mp3Path, mp3Data, 0644); err != nil {
		t.Fatalf("Failed to create test MP3: %v", err)
	}

	// Test data of various sizes
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "Small data",
			data: []byte("Secret message"),
		},
		{
			name: "Empty data",
			data: []byte{},
		},
		{
			name: "Binary data",
			data: []byte{0x00, 0x01, 0x02, 0x03},
		},
		{
			name: "Large data",
			data: make([]byte, 1024), // 1KB of data
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			outputPath := filepath.Join(tmpDir, "output.mp3")

			// Hide data
			err := HideInMP3(mp3Path, tc.data, outputPath)
			if err != nil {
				t.Fatalf("HideInMP3 failed: %v", err)
			}

			// Extract data
			extracted, err := ExtractFromMP3(outputPath)
			if err != nil {
				t.Fatalf("ExtractFromMP3 failed: %v", err)
			}

			// Compare
			if !bytes.Equal(extracted, tc.data) {
				t.Errorf("Extracted data doesn't match original")
			}

			// Verify MP3 structure
			outputData, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("Failed to read output MP3: %v", err)
			}

			// Check ID3v2 header is preserved
			if !bytes.HasPrefix(outputData, []byte("ID3")) {
				t.Error("ID3v2 header not preserved")
			}

			// Check MP3 frame is preserved
			if !bytes.Contains(outputData, []byte{0xFF, 0xFB, 0x90, 0x64}) {
				t.Error("MP3 frame not preserved")
			}

			// Check file is still valid MP3
			if len(outputData) < len(mp3Data) {
				t.Error("Output file is too small to be valid")
			}
		})
	}
}
