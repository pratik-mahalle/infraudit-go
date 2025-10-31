APP=api
PKG=./...
BIN=bin/$(APP)
DOCKER_IMAGE=infraaudit-go

.PHONY: build run clean docker-build docker-run deps fmt lint

build:
	GO111MODULE=on CGO_ENABLED=0 go build -o $(BIN) ./cmd/api

build-settings:
	GO111MODULE=on CGO_ENABLED=0 go build -o bin/settings ./settings

run:
	API_ADDR=:8080 DB_PATH=./data.db go run ./cmd/api

run-settings:
	go run ./settings

clean:
	rm -f $(BIN)

deps:
	go mod download

fmt:
	gofmt -s -w .

lint:
	go vet $(PKG)

docker-build:
	docker build -t $(DOCKER_IMAGE):latest .

PORT?=8080
DATA_DIR?=$(PWD)/data

docker-run:
	mkdir -p $(DATA_DIR)
	docker run --rm -p $(PORT):8080 -e API_ADDR=":8080" -e DB_PATH="/data/data.db" -v $(DATA_DIR):/data $(DOCKER_IMAGE):latest
