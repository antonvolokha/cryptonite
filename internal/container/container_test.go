package container

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// Helper function to create test files and return their paths and content map
func createTestFiles(t *testing.T) (string, map[string][]byte) {
	tmpDir := t.TempDir()

	testFiles := map[string][]byte{
		"file1.txt": []byte("Hello, World!"),
		"file2.bin": {0x00, 0x01, 0x02, 0x03},
		"large.dat": make([]byte, 1024*1024), // 1MB file
	}

	for name, content := range testFiles {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	return tmpDir, testFiles
}

// Helper function to verify container contents match expected files
func verifyContainerContents(t *testing.T, cont *Container, testFiles map[string][]byte) {
	if len(cont.Files) != len(testFiles) {
		t.Errorf("Wrong number of files. Got: %d, Want: %d", len(cont.Files), len(testFiles))
	}

	for _, file := range cont.Files {
		original, exists := testFiles[filepath.Base(file.Path)]
		if !exists {
			t.Errorf("Unexpected file: %s", file.Path)
			continue
		}

		if !bytes.Equal(file.Data, original) {
			t.Errorf("File content mismatch for %s", file.Path)
		}
	}
}

// Helper function to verify extracted files
func verifyExtractedFiles(t *testing.T, outputDir string, testFiles map[string][]byte) {
	for name, content := range testFiles {
		path := filepath.Join(outputDir, name)
		extracted, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read extracted file %s: %v", name, err)
			continue
		}

		if !bytes.Equal(extracted, content) {
			t.Errorf("Extracted content mismatch for %s", name)
		}
	}
}

// Test compressed container functionality
func TestCompressedContainer(t *testing.T) {
	tmpDir, testFiles := createTestFiles(t)

	// Create and fill container with compression enabled (default)
	cont := NewContainer()
	for name := range testFiles {
		path := filepath.Join(tmpDir, name)
		if err := cont.AddFile(path); err != nil {
			t.Fatalf("AddFile failed: %v", err)
		}
	}

	// Verify compression is enabled by default
	if !cont.UseCompression {
		t.Error("UseCompression should be true by default")
	}

	// Test serialization
	data := cont.Bytes()
	if len(data) == 0 {
		t.Error("Container serialization produced empty data")
	}

	// Make sure the magic number is present (first 2 bytes)
	if !bytes.Equal(data[0:2], MagicNumber) {
		t.Errorf("Missing magic number in container header")
	}

	// Make sure the version is correct (next 2 bytes)
	var version uint16
	if err := binary.Read(bytes.NewReader(data[2:4]), binary.LittleEndian, &version); err != nil {
		t.Fatalf("Failed to read version: %v", err)
	}
	if version != VersionCompressed {
		t.Errorf("Expected compressed version %d, got %d", VersionCompressed, version)
	}

	// Test deserialization
	newCont := NewContainer()
	if err := newCont.FromBytes(data); err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	// Verify UseCompression was set correctly during deserialization
	if !newCont.UseCompression {
		t.Error("UseCompression not preserved after deserialization")
	}

	// Verify content
	verifyContainerContents(t, newCont, testFiles)

	// Test extraction
	outputDir := filepath.Join(tmpDir, "compressed_output")
	if err := newCont.ExtractAll(outputDir); err != nil {
		t.Fatalf("ExtractAll failed: %v", err)
	}

	// Verify extracted files
	verifyExtractedFiles(t, outputDir, testFiles)
}

// Test uncompressed container functionality
func TestUncompressedContainer(t *testing.T) {
	tmpDir, testFiles := createTestFiles(t)

	// Create container with compression disabled
	cont := NewContainer()
	cont.UseCompression = false

	for name := range testFiles {
		path := filepath.Join(tmpDir, name)
		if err := cont.AddFile(path); err != nil {
			t.Fatalf("AddFile failed: %v", err)
		}
	}

	// Test serialization
	data := cont.Bytes()
	if len(data) == 0 {
		t.Error("Container serialization produced empty data")
	}

	// Make sure the magic number is present (first 2 bytes)
	if !bytes.Equal(data[0:2], MagicNumber) {
		t.Errorf("Missing magic number in container header")
	}

	// Make sure the version is correct (next 2 bytes)
	var version uint16
	if err := binary.Read(bytes.NewReader(data[2:4]), binary.LittleEndian, &version); err != nil {
		t.Fatalf("Failed to read version: %v", err)
	}
	if version != VersionUncompressed {
		t.Errorf("Expected uncompressed version %d, got %d", VersionUncompressed, version)
	}

	// Test deserialization
	newCont := NewContainer()
	if err := newCont.FromBytes(data); err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	// Verify UseCompression was set correctly during deserialization
	if newCont.UseCompression {
		t.Error("UseCompression not preserved after deserialization")
	}

	// Verify content
	verifyContainerContents(t, newCont, testFiles)

	// Test extraction
	outputDir := filepath.Join(tmpDir, "uncompressed_output")
	if err := newCont.ExtractAll(outputDir); err != nil {
		t.Fatalf("ExtractAll failed: %v", err)
	}

	// Verify extracted files
	verifyExtractedFiles(t, outputDir, testFiles)
}

// Test backward compatibility with legacy format (no version byte)
func TestLegacyFormatBackwardCompatibility(t *testing.T) {
	tmpDir, testFiles := createTestFiles(t)

	// Create a container to populate with test files
	cont := NewContainer()
	for name := range testFiles {
		path := filepath.Join(tmpDir, name)
		if err := cont.AddFile(path); err != nil {
			t.Fatalf("AddFile failed: %v", err)
		}
	}

	// Create legacy format by manually removing version byte
	// This simulates data from an older version without versioning/compression
	legacyFormatBytes := createLegacyFormatData(t, tmpDir, testFiles)

	// Test deserialization of legacy format
	legacyCont := NewContainer()
	if err := legacyCont.FromBytes(legacyFormatBytes); err != nil {
		t.Fatalf("FromBytes failed with legacy format: %v", err)
	}

	// Verify content
	verifyContainerContents(t, legacyCont, testFiles)

	// Test extraction
	outputDir := filepath.Join(tmpDir, "legacy_output")
	if err := legacyCont.ExtractAll(outputDir); err != nil {
		t.Fatalf("ExtractAll failed: %v", err)
	}

	// Verify extracted files
	verifyExtractedFiles(t, outputDir, testFiles)
}

// Create legacy format data (no version byte)
func createLegacyFormatData(t *testing.T, tmpDir string, testFiles map[string][]byte) []byte {
	buf := new(bytes.Buffer)

	// Write number of files
	if err := binary.Write(buf, binary.LittleEndian, int64(len(testFiles))); err != nil {
		t.Fatalf("Failed to write file count: %v", err)
	}

	for name, content := range testFiles {
		path := filepath.Join(tmpDir, name)

		// Write path length and path
		if err := binary.Write(buf, binary.LittleEndian, int64(len(path))); err != nil {
			t.Fatalf("Failed to write path length: %v", err)
		}
		buf.Write([]byte(path))

		// Write file size and data
		size := int64(len(content))
		if err := binary.Write(buf, binary.LittleEndian, size); err != nil {
			t.Fatalf("Failed to write file size: %v", err)
		}
		buf.Write(content)
	}

	return buf.Bytes()
}
