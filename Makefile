all: favpost

favpost: go build -o bin/favpost ./cmd/favpost/favpost.go
