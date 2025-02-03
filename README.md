# Cryptonite

Cryptonite is a CLI utility for secure file and folder encryption with the ability to hide data in MP3 files using steganography.

## Features

- 🔐 File and folder encryption using AES-256  
- 🎵 Steganography: hide encrypted data in MP3 files  
- 🔑 Secure password input  
- 📁 Support for both single files and entire directories  
- 🖥️ Cross-platform support (Windows, macOS, Linux)  

## Installation

### Using Pre-built Binaries

1. Download the latest release for your platform from [Releases](https://github.com/your-username/cryptonite/releases).  
2. Extract the archive.  
3. Add the binary to your `PATH` or use the full path to execute it.  

### Building from Source

```bash
# Clone the repository
git clone https://github.com/your-username/cryptonite.git
cd cryptonite

# Build the project
make build

# Install (optional)
sudo make install
```

## Usage

### Encrypt a File
```bash
cryptonite -encrypt -input /path/to/file -output encrypted.bin
```

### Decrypt a File
```bash
cryptonite -decrypt -input encrypted.bin -output /path/to/extract
```

### Encrypt and Hide Data in MP3
```bash
cryptonite -encrypt -input secret.doc -output hidden.mp3 -mp3 original.mp3
```

### Extract Data from MP3
```bash
cryptonite -decrypt -input hidden.mp3 -mp3 yes -output extracted.doc
```

### Run Tests
```bash
go test -v ./...
```

### Lint the Code
```bash
golangci-lint run
```

## Project Structure

```
.
├── cmd/
│   └── main.go            # Entry point
├── internal/
│   ├── crypto/            # Encryption logic
│   ├── container/         # Container operations
│   └── steganography/     # Steganography operations
├── tests/                 # Test scripts
└── build/                 # Compiled binaries
```

## Building for Different Platforms

### Build for All Platforms
```bash
make build
```

### Build for a Specific Platform
```bash
GOOS=darwin GOARCH=arm64 go build -o build/cryptonite-darwin-arm64 ./cmd/main.go
```

## Troubleshooting

### 1. Encryption/Decryption Errors  
   - Verify the correct password and encryption key.  
   - Check that the input file is accessible.  

### 2. MP3 Steganography Issues  
   - Ensure the original MP3 file is valid.  
   - Check file size limitations.  
   - Verify MP3 format compatibility.  

### 3. Memory Issues  
   - Use smaller chunks for large files.  
   - Ensure sufficient system resources.  

## Contact

If you have questions or suggestions:  
1. Open an **Issue** on GitHub.  
2. Submit a **Pull Request**.  
3. Contact the maintainers.  

## Acknowledgments

- [golang/go](https://github.com/golang/go)  
- [All contributors](https://github.com/your-username/cryptonite/graphs/contributors)  

