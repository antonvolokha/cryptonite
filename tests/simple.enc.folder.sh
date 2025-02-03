PASSWORD="mysecretpass"
CONTAINER_NAME="container.enc"

go run ../cmd/main.go -encrypt -input ./data -output $CONTAINER_NAME -password $PASSWORD
go run ../cmd/main.go -decrypt -input container.enc -output ./results -password $PASSWORD