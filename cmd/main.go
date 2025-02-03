package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/term"

	"cryptoutils/internal/container"
	"cryptoutils/internal/crypto"
	"cryptoutils/internal/steganography"
)

func getPassword() string {
	fmt.Print("Enter password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("Could not read password:", err)
	}
	fmt.Println() // Add newline after password input
	return string(password)
}

func main() {
	encrypt := flag.Bool("encrypt", false, "Encrypt files")
	decrypt := flag.Bool("decrypt", false, "Decrypt container")
	input := flag.String("input", "", "Input file or directory")
	output := flag.String("output", "", "Output container file")
	mp3 := flag.String("mp3", "", "MP3 file to hide container in")

	flag.Parse()

	if *input == "" {
		log.Fatal("Input path is required")
	}

	// Get password securely
	password := getPassword()
	if password == "" {
		log.Fatal("Password is required")
	}

	if *encrypt {
		// Create crypto container
		cont := container.NewContainer()

		// Add files to container
		err := filepath.Walk(*input, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return cont.AddFile(path)
			}
			return nil
		})

		if err != nil {
			log.Fatal(err)
		}

		// Encrypt container
		encrypted, err := crypto.Encrypt(cont.Bytes(), password)
		if err != nil {
			log.Fatal(err)
		}

		if *mp3 != "" {
			// Hide in MP3
			err = steganography.HideInMP3(*mp3, encrypted, *output)
		} else {
			// Save as regular container
			err = os.WriteFile(*output, encrypted, 0644)
		}

		if err != nil {
			log.Fatal(err)
		}

	} else if *decrypt {
		var encrypted []byte
		var err error

		if *mp3 != "" {
			// Extract from MP3
			encrypted, err = steganography.ExtractFromMP3(*input)
		} else {
			// Read regular container
			encrypted, err = os.ReadFile(*input)
		}

		if err != nil {
			log.Fatal(err)
		}

		// Decrypt container
		decrypted, err := crypto.Decrypt(encrypted, password)
		if err != nil {
			log.Fatal(err)
		}

		// Extract files
		cont := container.NewContainer()
		err = cont.FromBytes(decrypted)
		if err != nil {
			log.Fatal(err)
		}

		err = cont.ExtractAll(*output)
		if err != nil {
			log.Fatal(err)
		}
	}
}
