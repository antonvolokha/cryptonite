package container

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
	"github.com/schollz/progressbar/v3"
)

type FileEntry struct {
	Path string
	Size int64
	Data []byte
}

// Version constants to ensure backward compatibility
const (
	VersionUncompressed byte = 0
	VersionCompressed   byte = 1
	CurrentVersion      byte = VersionCompressed
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

	// Write format version
	version := VersionUncompressed
	if c.UseCompression {
		version = VersionCompressed
	}
	buf.WriteByte(version)

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

	// If compression is enabled, compress the content
	if c.UseCompression {
		encoder, err := zstd.NewWriter(buf)
		if err != nil {
			panic(err)
		}
		if _, err := encoder.Write(contentBuf.Bytes()); err != nil {
			panic(err)
		}
		if err := encoder.Close(); err != nil {
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

	// Read and check version
	version := data[0]
	data = data[1:] // Remove version byte

	// Use a reader for content based on version
	var contentReader io.Reader

	switch version {
	case VersionUncompressed:
		// For uncompressed data, just read directly
		contentReader = bytes.NewReader(data)
		c.UseCompression = false

	case VersionCompressed:
		// For compressed data, set up a zstd decoder
		decoder, err := zstd.NewReader(bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to create zstd decoder: %w", err)
		}
		defer decoder.Close()
		contentReader = decoder
		c.UseCompression = true

	default:
		// For unknown version, try to handle as legacy uncompressed format
		// (no version byte at beginning)
		contentReader = bytes.NewReader(append([]byte{version}, data...))
		c.UseCompression = false
	}

	// Read number of files
	var numFiles int64
	if err := binary.Read(contentReader, binary.LittleEndian, &numFiles); err != nil {
		return fmt.Errorf("failed to read number of files: %w", err)
	}

	c.Files = make([]FileEntry, 0, numFiles)

	for i := int64(0); i < numFiles; i++ {
		// Read path length
		var pathLen int64
		if err := binary.Read(contentReader, binary.LittleEndian, &pathLen); err != nil {
			return fmt.Errorf("failed to read path length for file %d: %w", i, err)
		}

		// Read path
		pathBytes := make([]byte, pathLen)
		if _, err := io.ReadFull(contentReader, pathBytes); err != nil {
			return fmt.Errorf("failed to read path for file %d: %w", i, err)
		}

		// Read file size
		var size int64
		if err := binary.Read(contentReader, binary.LittleEndian, &size); err != nil {
			return fmt.Errorf("failed to read size for file %d: %w", i, err)
		}

		// Read file data
		data := make([]byte, size)
		if _, err := io.ReadFull(contentReader, data); err != nil {
			return fmt.Errorf("failed to read data for file %d: %w", i, err)
		}

		c.Files = append(c.Files, FileEntry{
			Path: string(pathBytes),
			Size: size,
			Data: data,
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
