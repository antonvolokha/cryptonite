#!/bin/bash
HIDDEN_FILE="hidden.mp3"

# Encrypt and hide in MP3
printf "mysecretpass" | go run ../cmd/main.go -encrypt -input ./data -output $HIDDEN_FILE -mp3 ./data/original.mp3

# Extract and decrypt
printf "mysecretpass" | go run ../cmd/main.go -decrypt -input $HIDDEN_FILE -mp3 yes -output ./results/folder
