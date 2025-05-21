package container

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"github.com/ulikunitz/xz/lzma"
)

type FileEntry struct {
	Path string
	Size int64
	Data []byte
}

// Magic number for identifying container files (FC = FileCrypt)
var MagicNumber = []byte{0x46, 0x43} // "FC" in ASCII

// Version constants to ensure backward compatibility
const (
	VersionLegacy       uint16 = 0 // For legacy format detection (no version header)
	VersionUncompressed uint16 = 1 // Uncompressed format
	VersionCompressed   uint16 = 2 // LZMA2 compressed format
	CurrentVersion      uint16 = VersionCompressed
)

type Container struct {
	Files          []FileEntry
	UseCompression bool // whether to use compression when serializing
}

func NewContainer() *Container {
	return &Container{
		Files:          make([]FileEntry, 0),
		UseCompression: true, // Enable compression by default for new containers
	}
}

func (c *Container) AddFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	c.Files = append(c.Files, FileEntry{
		Path: path,
		Size: int64(len(data)),
		Data: data,
	})

	return nil
}

func (c *Container) Bytes() []byte {
	buf := new(bytes.Buffer)

	// Write container header (magic number + version)
	buf.Write(MagicNumber)
	var formatVersion = VersionUncompressed
	if c.UseCompression {
		formatVersion = VersionCompressed
	}
	if err := binary.Write(buf, binary.LittleEndian, formatVersion); err != nil {
		panic(err)
	}

	// Create content buffer for all files
	contentBuf := new(bytes.Buffer)

	// Write number of files
	if err := binary.Write(contentBuf, binary.LittleEndian, int64(len(c.Files))); err != nil {
		// Since this is writing to a bytes.Buffer, errors are not expected
		// but we should handle them anyway
		panic(err)
	}

	for _, file := range c.Files {
		// Write path length and path
		if err := binary.Write(contentBuf, binary.LittleEndian, int64(len(file.Path))); err != nil {
			panic(err)
		}
		contentBuf.Write([]byte(file.Path))

		// Write file size and data
		if err := binary.Write(contentBuf, binary.LittleEndian, file.Size); err != nil {
			panic(err)
		}
		contentBuf.Write(file.Data)
	}

	// Note: Magic number and version are already written at the beginning of this method

	// Now write the content (possibly compressed)
	if c.UseCompression {
		// Create LZMA2 writer with default settings
		w, err := lzma.NewWriter2(buf)
		if err != nil {
			panic(err)
		}

		if _, err := w.Write(contentBuf.Bytes()); err != nil {
			panic(err)
		}

		if err := w.Close(); err != nil {
			panic(err)
		}
	} else {
		// Write uncompressed content
		buf.Write(contentBuf.Bytes())
	}

	return buf.Bytes()
}

func (c *Container) FromBytes(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("empty data")
	}

	// Check if this is a versioned format with our magic number (at least 4 bytes for header)
	if len(data) >= 4 && bytes.Equal(data[0:2], MagicNumber) {
		// This appears to be a versioned format
		// Read version number (2 bytes, little endian)
		var version uint16
		if err := binary.Read(bytes.NewReader(data[2:4]), binary.LittleEndian, &version); err != nil {
			return fmt.Errorf("failed to read version number: %w", err)
		}

		fmt.Printf("Debug: Detected versioned format: %d\n", version)

		// Handle versioned format based on version
		dataWithoutHeader := data[4:] // Remove 4-byte header

		switch version {
		case VersionUncompressed:
			// For uncompressed format
			return c.readUncompressedFormat(dataWithoutHeader)

		case VersionCompressed:
			// For compressed format with LZMA2
			return c.readCompressedFormat(dataWithoutHeader)

		default:
			return fmt.Errorf("unknown version number: %d", version)
		}
	}

	// If versioned format didn't work, try legacy format as a fallback
	err := c.readLegacyFormat(data)
	if err == nil {
		// Successfully read as legacy format
		return nil
	}

	// If we get here, neither format worked
	return fmt.Errorf("could not parse container data: not a valid container format: %v", err)
}

// readUncompressedFormat reads a container in the versioned uncompressed format
func (c *Container) readUncompressedFormat(data []byte) error {
	c.UseCompression = false

	// Create a reader for the data
	buf := bytes.NewReader(data)

	// Read container content
	return c.readContainerData(buf)
}

// readCompressedFormat reads a container in the versioned compressed format
func (c *Container) readCompressedFormat(data []byte) error {
	// Create LZMA2 reader
	reader, err := lzma.NewReader2(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create LZMA2 decoder: %w", err)
	}

	c.UseCompression = true

	// Read container content
	return c.readContainerData(reader)
}

// readLegacyFormat reads a container in the legacy format (no version byte)
func (c *Container) readLegacyFormat(data []byte) error {
	// Legacy format doesn't use compression
	c.UseCompression = false

	// Create a reader for the data
	buf := bytes.NewReader(data)

	// Read container content
	return c.readContainerData(buf)
}

// readContainerData reads a container from the given reader
func (c *Container) readContainerData(reader io.Reader) error {
	// Read number of files
	var numFiles int64
	if err := binary.Read(reader, binary.LittleEndian, &numFiles); err != nil {
		return fmt.Errorf("failed to read number of files: %w", err)
	}

	// Validate numFiles to prevent panic with invalid data and do a sanity check
	if numFiles <= 0 || numFiles > 10000 { // Reasonable limit for number of files
		return fmt.Errorf("invalid or unlikely number of files: %d", numFiles)
	}

	// Debug output
	fmt.Printf("Debug: Processing container with %d files\n", numFiles)

	c.Files = make([]FileEntry, 0, numFiles)

	for i := int64(0); i < numFiles; i++ {
		// Read path length
		var pathLen int64
		if err := binary.Read(reader, binary.LittleEndian, &pathLen); err != nil {
			return fmt.Errorf("failed to read path length for file %d: %w", i, err)
		}

		// Validate path length
		if pathLen < 0 || pathLen > 32768 { // Max 32KB path length
			return fmt.Errorf("invalid path length for file %d: %d", i, pathLen)
		}

		// Read path
		pathBytes := make([]byte, pathLen)
		if _, err := io.ReadFull(reader, pathBytes); err != nil {
			return fmt.Errorf("failed to read path for file %d: %w", i, err)
		}

		// Debug path
		filePath := string(pathBytes)
		fmt.Printf("Debug: Reading file %d: %s\n", i, filePath)

		// Read file size
		var size int64
		if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
			return fmt.Errorf("failed to read size for file %d: %w", i, err)
		}

		// Validate file size
		if size < 0 || size > 1024*1024*1024 { // Max 1GB file size
			return fmt.Errorf("invalid file size for file %d: %d", i, size)
		}

		// Debug size
		fmt.Printf("Debug: File %d size: %d bytes\n", i, size)

		// Read file data
		fileData := make([]byte, size)
		if _, err := io.ReadFull(reader, fileData); err != nil {
			return fmt.Errorf("failed to read data for file %d: %w", i, err)
		}

		c.Files = append(c.Files, FileEntry{
			Path: filePath,
			Size: size,
			Data: fileData,
		})
	}

	return nil
}

func (c *Container) ExtractAll(outputDir string) error {
	for _, file := range c.Files {
		fullPath := filepath.Join(outputDir, filepath.Base(file.Path))

		// Create directories if needed
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}

		// Write file
		if err := os.WriteFile(fullPath, file.Data, 0644); err != nil {
			return err
		}
	}

	return nil
}

func (c *Container) ExtractAllWithProgress(outputDir string, bar *progressbar.ProgressBar) error {
	for _, file := range c.Files {
		fullPath := filepath.Join(outputDir, filepath.Base(file.Path))

		// Create directories if needed
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Write file
		if err := os.WriteFile(fullPath, file.Data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		// Update progress bar
		if err := bar.Add(1); err != nil {
			return fmt.Errorf("failed to update progress: %w", err)
		}
	}

	return nil
}
