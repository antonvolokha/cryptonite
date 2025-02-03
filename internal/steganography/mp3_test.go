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

	// Create a minimal valid MP3 file
	mp3Data := []byte{
		// ID3v2 header
		0x49, 0x44, 0x33, // "ID3"
		0x03, 0x00, // version
		0x00,                   // flags
		0x00, 0x00, 0x00, 0x00, // size
		// MP3 frame
		0xFF, 0xFB, 0x90, 0x64, // MPEG1 Layer3
		// Some audio data
		0x01, 0x02, 0x03, 0x04,
	}

	mp3Path := filepath.Join(tmpDir, "test.mp3")
	if err := os.WriteFile(mp3Path, mp3Data, 0644); err != nil {
		t.Fatalf("Failed to create test MP3: %v", err)
	}

	// Test data to hide
	testData := []byte("Secret message")

	// Hide data
	outputPath := filepath.Join(tmpDir, "output.mp3")
	err := HideInMP3(mp3Path, testData, outputPath)
	if err != nil {
		t.Fatalf("HideInMP3 failed: %v", err)
	}

	// Extract data
	extracted, err := ExtractFromMP3(outputPath)
	if err != nil {
		t.Fatalf("ExtractFromMP3 failed: %v", err)
	}

	// Compare
	if !bytes.Equal(extracted, testData) {
		t.Errorf("Extracted data doesn't match original.\nGot: %v\nWant: %v", extracted, testData)
	}

	// Verify MP3 is still valid
	outputData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output MP3: %v", err)
	}

	// Check ID3 header is preserved
	if !bytes.HasPrefix(outputData, []byte("ID3")) {
		t.Error("ID3 header not preserved")
	}

	// Check MP3 frame is preserved
	if !bytes.Contains(outputData, []byte{0xFF, 0xFB, 0x90, 0x64}) {
		t.Error("MP3 frame not preserved")
	}
}
