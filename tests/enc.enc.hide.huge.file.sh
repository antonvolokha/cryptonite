PASSWORD="mysecretpass"
HIDDEN_FILE="hidden.mp3"
FILE_NAME="secret_video.mp4"

go run ../cmd/main.go -encrypt -input ./data/$FILE_NAME -output $HIDDEN_FILE -mp3 ./data/original.mp3 -password $PASSWORD
go run ../cmd/main.go -decrypt -input $HIDDEN_FILE -mp3 yes -output ./results/$FILE_NAME -password $PASSWORD
