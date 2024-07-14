default: install

install:
	go build -o=./cmd/bast ./cmd/bast
	go install ./cmd/bast