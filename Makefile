-include .env
export

.DEFAULT_GOAL := run
PROJECT := climber-count
SRC := $(wildcard *.go)

$(PROJECT): $(SRC)
	go build -o $@ ./...

.PHONY: build
build: $(PROJECT)

.PHONY: test
test:
	go test -v ./...

.PHONY: run
run: $(PROJECT)
	./$<

.PHONY: clean
clean:
	go clean
