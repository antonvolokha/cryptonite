package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"

	"cryptoutils/internal/container"
	"cryptoutils/internal/crypto"
	"cryptoutils/internal/steganography"
)

const asciiArt = `
   ______                 __             _ __    
  / ____/______  ______  / /_____  ____(_) /__  
 / /   / ___/ / / / __ \/ __/ __ \/ __ \/ / _ \ 
/ /___/ /  / /_/ / /_/ / /_/ /_/ / / / / /  __/ 
\____/_/   \__, / .___/\__/\____/_/ /_/_/\___/  
          /____/_/                               
`

const version = "v0.0.3"

func showHelp() {
	fmt.Printf("%s\nVersion: %s\n\n", asciiArt, version)
	fmt.Print("Usage:\n")
	fmt.Print("  cryptonite [options]\n")
	fmt.Print("\nOptions:\n")
	flag.PrintDefaults()
	fmt.Print("\nExamples:\n")
	fmt.Print("  Encrypt file:\n")
	fmt.Print("    cryptonite -encrypt -input secret.doc -output encrypted.bin\n")
	fmt.Print("\n  Decrypt file:\n")
	fmt.Print("    cryptonite -decrypt -input encrypted.bin -output secret.doc\n")
	fmt.Print("\n  Hide in MP3:\n")
	fmt.Print("    cryptonite -encrypt -input secret.doc -output hidden.mp3 -mp3 original.mp3\n")
	fmt.Print("\n  Extract from MP3:\n")
	fmt.Print("    cryptonite -decrypt -input hidden.mp3 -mp3 yes -output secret.doc\n")
}

func getPassword() string {
	// Try reading from pipe first
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped in
		var password string
		if _, err := fmt.Scanln(&password); err != nil {
			log.Fatal("Failed to read password:", err)
		}
		return password
	}

	// No pipe, read from terminal
	fmt.Print("Enter password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("Could not read password:", err)
	}
	fmt.Print("\n")
	return string(password)
}

func promptString(prompt string) string {
	var input string
	fmt.Print(prompt + ": ")
	if _, err := fmt.Scanln(&input); err != nil {
		log.Fatal("Failed to read input:", err)
	}
	return input
}

func promptChoice(prompt string, options []string) int {
	fmt.Print(prompt + "\n")
	for i, opt := range options {
		fmt.Printf("%d) %s\n", i+1, opt)
	}
	var choice int
	for {
		fmt.Print("Enter your choice (1-" + fmt.Sprint(len(options)) + "): ")
		if _, err := fmt.Scanln(&choice); err != nil {
			fmt.Print("Invalid input, please try again\n")
			continue
		}
		if choice > 0 && choice <= len(options) {
			return choice - 1
		}
		fmt.Print("Invalid choice, please try again\n")
	}
}

func main() {
	help := flag.Bool("help", false, "Show help message")
	encrypt := flag.Bool("encrypt", false, "Encrypt files")
	decrypt := flag.Bool("decrypt", false, "Decrypt container")
	input := flag.String("input", "", "Input file or directory")
	output := flag.String("output", "", "Output container file")
	mp3 := flag.String("mp3", "", "MP3 file to hide container in")

	flag.Parse()

	// Show ASCII art header
	fmt.Print(asciiArt)
	fmt.Printf("Version: %s\n\n", version)

	if *help {
		showHelp()
		return
	}

	// Interactive mode if no flags provided
	if flag.NFlag() == 0 {
		options := []string{"Encrypt files", "Decrypt files"}
		choice := promptChoice("What would you like to do?", options)

		*encrypt = choice == 0
		*decrypt = choice == 1

		*input = promptString("Enter input path")
		*output = promptString("Enter output path")

		mp3Options := []string{"No", "Yes"}
		if promptChoice("Use MP3 steganography?", mp3Options) == 1 {
			if *encrypt {
				*mp3 = promptString("Enter original MP3 path")
			} else {
				*mp3 = "yes"
			}
		}
	}

	if *input == "" {
		log.Fatal("Input path is required")
	}

	if *output == "" {
		log.Fatal("Output path is required")
	}

	// Get password securely
	password := getPassword()
	if password == "" {
		log.Fatal("Password is required")
	}

	if *encrypt {
		// Create crypto container
		cont := container.NewContainer()

		// Create progress bar for walking files
		var totalFiles int
		if err := filepath.Walk(*input, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				totalFiles++
			}
			return nil
		}); err != nil {
			log.Fatal("Failed to count files:", err)
		}

		bar := progressbar.Default(int64(totalFiles), "Processing files")

		// Add files to container
		err := filepath.Walk(*input, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				if err := cont.AddFile(path); err != nil {
					return err
				}
				if err := bar.Add(1); err != nil {
					return fmt.Errorf("failed to update progress: %w", err)
				}
			}
			return nil
		})

		if err != nil {
			log.Fatal(err)
		}

		fmt.Print("\nEncrypting data...\n")
		bar = progressbar.Default(1, "Encrypting")

		// Encrypt container
		encrypted, err := crypto.Encrypt(cont.Bytes(), password)
		if err != nil {
			log.Fatal(err)
		}
		if err := bar.Add(1); err != nil {
			log.Fatal("Failed to update progress:", err)
		}

		fmt.Print("\nSaving result...\n")
		bar = progressbar.Default(1, "Saving")

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
		if err := bar.Add(1); err != nil {
			log.Fatal("Failed to update progress:", err)
		}

		fmt.Print("\nOperation completed successfully!\n")

	} else if *decrypt {
		fmt.Print("Reading encrypted data...\n")
		bar := progressbar.Default(1, "Reading")

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
		if err := bar.Add(1); err != nil {
			log.Fatal("Failed to update progress:", err)
		}

		fmt.Print("\nDecrypting data...\n")
		bar = progressbar.Default(1, "Decrypting")

		// Decrypt container
		decrypted, err := crypto.Decrypt(encrypted, password)
		if err != nil {
			log.Fatal(err)
		}
		if err := bar.Add(1); err != nil {
			log.Fatal("Failed to update progress:", err)
		}

		fmt.Print("\nExtracting files...\n")

		// Extract files
		cont := container.NewContainer()
		err = cont.FromBytes(decrypted)
		if err != nil {
			log.Fatal(err)
		}

		bar = progressbar.Default(int64(len(cont.Files)), "Extracting")
		err = cont.ExtractAllWithProgress(*output, bar)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Print("\nOperation completed successfully!\n")
	}
}
