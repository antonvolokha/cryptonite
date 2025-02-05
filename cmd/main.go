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

const version = "v0.0.1"

func showHelp() {
	fmt.Printf("%s\nVersion: %s\n\n", asciiArt, version)
	fmt.Println("Usage:")
	fmt.Println("  cryptonite [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Println("  Encrypt file:")
	fmt.Println("    cryptonite -encrypt -input secret.doc -output encrypted.bin")
	fmt.Println("\n  Decrypt file:")
	fmt.Println("    cryptonite -decrypt -input encrypted.bin -output secret.doc")
	fmt.Println("\n  Hide in MP3:")
	fmt.Println("    cryptonite -encrypt -input secret.doc -output hidden.mp3 -mp3 original.mp3")
	fmt.Println("\n  Extract from MP3:")
	fmt.Println("    cryptonite -decrypt -input hidden.mp3 -mp3 yes -output secret.doc")
}

func getPassword() string {
	// Try reading from pipe first
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped in
		var password string
		fmt.Scanln(&password)
		return password
	}

	// No pipe, read from terminal
	fmt.Print("Enter password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal("Could not read password:", err)
	}
	fmt.Println() // Add newline after password input
	return string(password)
}

func promptString(prompt string) string {
	var input string
	fmt.Print(prompt + ": ")
	fmt.Scanln(&input)
	return input
}

func promptChoice(prompt string, options []string) int {
	fmt.Println(prompt)
	for i, opt := range options {
		fmt.Printf("%d) %s\n", i+1, opt)
	}
	var choice int
	for {
		fmt.Print("Enter your choice (1-" + fmt.Sprint(len(options)) + "): ")
		fmt.Scanln(&choice)
		if choice > 0 && choice <= len(options) {
			return choice - 1
		}
		fmt.Println("Invalid choice, please try again")
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
	fmt.Println(asciiArt)
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
		filepath.Walk(*input, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				totalFiles++
			}
			return nil
		})

		bar := progressbar.Default(int64(totalFiles), "Processing files")

		// Add files to container
		err := filepath.Walk(*input, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				err = cont.AddFile(path)
				bar.Add(1)
				return err
			}
			return nil
		})

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("\nEncrypting data...")
		bar = progressbar.Default(1, "Encrypting")

		// Encrypt container
		encrypted, err := crypto.Encrypt(cont.Bytes(), password)
		if err != nil {
			log.Fatal(err)
		}
		bar.Add(1)

		fmt.Println("\nSaving result...")
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
		bar.Add(1)

		fmt.Println("\nOperation completed successfully!")

	} else if *decrypt {
		fmt.Println("Reading encrypted data...")
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
		bar.Add(1)

		fmt.Println("\nDecrypting data...")
		bar = progressbar.Default(1, "Decrypting")

		// Decrypt container
		decrypted, err := crypto.Decrypt(encrypted, password)
		if err != nil {
			log.Fatal(err)
		}
		bar.Add(1)

		fmt.Println("\nExtracting files...")

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

		fmt.Println("\nOperation completed successfully!")
	}
}
