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

.PHONY: image
image: export GOOS=linux
image: export GOARCH=arm64
image:
	go build -o $(PROJECT) ./...
	docker buildx build -t ghcr.io/eiri/$(PROJECT):latest . --platform=linux/arm64

.PHONY: docker-up
docker-up:
	docker compose up -d

.PHONY: docker-down
docker-down:
	docker compose down

.PHONY: docker-logs
docker-logs:
	docker compose logs
