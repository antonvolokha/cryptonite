PASSWORD="mysecretpass"
CONTAINER_NAME="container.enc"
FILE_NAME="very_secret_doc.pdf"

go run ../cmd/main.go -encrypt -input ./data/$FILE_NAME -output $CONTAINER_NAME -password $PASSWORD
go run ../cmd/main.go -decrypt -input container.enc -output ./results/$FILE_NAME -password $PASSWORD