#!/bin/bash
CONTAINER_NAME="container.enc"
FILE_NAME="very_secret_doc.pdf"

# Encrypt
printf "mysecretpass" | go run ../cmd/main.go -encrypt -input ./data/$FILE_NAME -output $CONTAINER_NAME

# Decrypt
printf "mysecretpass" | go run ../cmd/main.go -decrypt -input $CONTAINER_NAME -output ./results/$FILE_NAME