package container

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestContainer(t *testing.T) {
	// Create temporary test files
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

	// Create and fill container
	cont := NewContainer()
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

	// Test deserialization
	newCont := NewContainer()
	if err := newCont.FromBytes(data); err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	// Compare files
	if len(newCont.Files) != len(testFiles) {
		t.Errorf("Wrong number of files. Got: %d, Want: %d", len(newCont.Files), len(testFiles))
	}

	for _, file := range newCont.Files {
		original, exists := testFiles[filepath.Base(file.Path)]
		if !exists {
			t.Errorf("Unexpected file: %s", file.Path)
			continue
		}

		if !bytes.Equal(file.Data, original) {
			t.Errorf("File content mismatch for %s", file.Path)
		}
	}

	// Test extraction
	outputDir := filepath.Join(tmpDir, "output")
	if err := newCont.ExtractAll(outputDir); err != nil {
		t.Fatalf("ExtractAll failed: %v", err)
	}

	// Verify extracted files
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
