-include .env
export

.DEFAULT_GOAL := run
PROJECT := climber-count
SRC := $(wildcard *.go)

$(PROJECT): $(SRC)
	go build -o $@ ./...

occupancy.html:
	curl 'https://portal.rockgympro.com/portal/public/$(CC_UID)/occupancy?iframeid=occupancyCounter&fId=$(CC_FID)' -s -S -v -H "Accept: application/json, */*" -o $@

.PHONY: build
build: $(PROJECT)

.PHONY: run
run: $(PROJECT) occupancy.html
	./$< -gym SBL

.PHONY: clean
clean:
	go clean
