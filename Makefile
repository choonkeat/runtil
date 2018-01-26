SRC = $(shell find . -iname '*.go')

build/runtil: $(SRC)
	go build -o build/runtil *.go

test: $(SRC)
	go test ./...
