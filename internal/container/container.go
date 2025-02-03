package container

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
)

type FileEntry struct {
	Path string
	Size int64
	Data []byte
}

type Container struct {
	Files []FileEntry
}

func NewContainer() *Container {
	return &Container{
		Files: make([]FileEntry, 0),
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

	// Write number of files
	binary.Write(buf, binary.LittleEndian, int64(len(c.Files)))

	for _, file := range c.Files {
		// Write path length and path
		binary.Write(buf, binary.LittleEndian, int64(len(file.Path)))
		buf.Write([]byte(file.Path))

		// Write file size and data
		binary.Write(buf, binary.LittleEndian, file.Size)
		buf.Write(file.Data)
	}

	return buf.Bytes()
}

func (c *Container) FromBytes(data []byte) error {
	buf := bytes.NewReader(data)

	// Read number of files
	var numFiles int64
	if err := binary.Read(buf, binary.LittleEndian, &numFiles); err != nil {
		return err
	}

	c.Files = make([]FileEntry, 0, numFiles)

	for i := int64(0); i < numFiles; i++ {
		// Read path length
		var pathLen int64
		if err := binary.Read(buf, binary.LittleEndian, &pathLen); err != nil {
			return err
		}

		// Read path
		pathBytes := make([]byte, pathLen)
		if _, err := io.ReadFull(buf, pathBytes); err != nil {
			return err
		}

		// Read file size
		var size int64
		if err := binary.Read(buf, binary.LittleEndian, &size); err != nil {
			return err
		}

		// Read file data
		data := make([]byte, size)
		if _, err := io.ReadFull(buf, data); err != nil {
			return err
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