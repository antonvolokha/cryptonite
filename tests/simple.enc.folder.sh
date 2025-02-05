#!/bin/bash
CONTAINER_NAME="container.enc"

# Encrypt
printf "mysecretpass" | go run ../cmd/main.go -encrypt -input ./data -output $CONTAINER_NAME

# Decrypt
printf "mysecretpass" | go run ../cmd/main.go -decrypt -input $CONTAINER_NAME -output ./results