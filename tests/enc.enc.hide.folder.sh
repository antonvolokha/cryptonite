PASSWORD="mysecretpass"
HIDDEN_FILE="hidden.mp3"

go run ../cmd/main.go -encrypt -input ./data -output $HIDDEN_FILE -mp3 ./data/original.mp3 -password $PASSWORD
go run ../cmd/main.go -decrypt -input $HIDDEN_FILE -mp3 yes -output ./results/folder -password $PASSWORD
