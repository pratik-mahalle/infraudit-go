APP=api
PKG=./...
BIN=bin/$(APP)
DOCKER_IMAGE=infraaudit-go

.PHONY: build run clean docker-build docker-run deps fmt lint

build:
	GO111MODULE=on CGO_ENABLED=0 go build -o $(BIN) ./cmd/api

run:
	API_ADDR=:5000 DB_PATH=./data.db go run ./cmd/api

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

PORT?=5000
DATA_DIR?=$(PWD)/data

docker-run:
	mkdir -p $(DATA_DIR)
	docker run --rm -p $(PORT):5000 -e API_ADDR=":5000" -e DB_PATH="/data/data.db" -v $(DATA_DIR):/data $(DOCKER_IMAGE):latest
