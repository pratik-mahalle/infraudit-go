APP=api
PKG=./...
BIN=bin/$(APP)
DOCKER_IMAGE=infraaudit-go

.PHONY: build build-cli install-cli run clean docker-build docker-run deps fmt lint swagger swagger-install

build:
	GO111MODULE=on CGO_ENABLED=0 go build -o $(BIN) ./cmd/api

build-cli:
	GO111MODULE=on CGO_ENABLED=0 go build -o bin/infraaudit ./cmd/cli

install-cli:
	go install ./cmd/cli

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

# Swagger documentation
swagger-install:
	@which swag > /dev/null || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)

swagger: swagger-install
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
	@echo "Swagger docs generated successfully!"

swagger-validate: swagger-install
	@echo "Validating Swagger documentation is up to date..."
	@swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
	@git diff --exit-code docs/ || (echo "ERROR: Swagger docs are out of date. Run 'make swagger' and commit the changes." && exit 1)
	@echo "Swagger docs are up to date!"
