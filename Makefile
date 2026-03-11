APP := coc

.PHONY: run init test install build

SLOT ?= slot1

run:
	go run ./cmd/$(APP) $(SLOT)

init:
	go run ./cmd/$(APP) init $(SLOT)

test:
	go test ./...

install:
	go install ./cmd/$(APP)

build:
	mkdir -p bin
	go build -o bin/$(APP) ./cmd/$(APP)