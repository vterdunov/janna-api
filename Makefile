PROG_NAME = janna-api
IMAGE_NAME = vterdunov/$(PROG_NAME)

PORT ?= 8080
COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%S')
PROJECT ?= github.com/vterdunov/${PROG_NAME}

GO_VARS=CGO_ENABLED=0 GOOS=linux GOARCH=amd64
GO_LDFLAGS :="
GO_LDFLAGS += -s -w
GO_LDFLAGS += -X ${PROJECT}/internal/version.Commit=${COMMIT}
GO_LDFLAGS += -X ${PROJECT}/internal/version.BuildTime=${BUILD_TIME}
GO_LDFLAGS +="

TAG ?= $(COMMIT)

GOLANGCI_LINTER_VERSION = v1.10.1
OPENAPI_GENERATOR_CLI_VERSION = v3.2.2

all: lint test api-doc-convert docker

.PHONY: docker
docker:
	docker build --tag=$(IMAGE_NAME):$(COMMIT) --tag=$(IMAGE_NAME):latest --file build/Dockerfile .

.PHONY: push
push:
	docker tag $(IMAGE_NAME):$(COMMIT) $(IMAGE_NAME):$(TAG)
	docker push $(IMAGE_NAME):$(TAG)

.PHONY: dep
dep:
	@dep ensure -v

.PHONY: compile
compile: clean
	 $(GO_VARS) go build -v -ldflags $(GO_LDFLAGS) -o $(PROG_NAME) ./cmd/janna/server.go

.PHONY: cgo-compile
cgo-compile: clean
	 go build -v -o $(PROG_NAME) ./cmd/janna/server.go

.PHONY: start
start:
	@env `cat .env | grep -v ^# | xargs` go run -race ./cmd/janna/server.go

start-binary: compile
	@env `cat .env | grep -v ^# | xargs` ./janna-api

.PHONY: dc
dc: dc-clean
	docker-compose -f deploy/docker-compose.dev.yml up --build

dc-clean:
	docker-compose -f deploy/docker-compose.dev.yml down --volumes

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	@echo Linting...
	@docker run -it --rm -v $(CURDIR):/go/src/$(PROJECT) -w /go/src/$(PROJECT) golangci/golangci-lint:$(GOLANGCI_LINTER_VERSION) run

.PHONY: clean
clean:
	@rm -f ${PROG_NAME}

api-doc-validate:
	@docker run \
		--rm \
		--entrypoint='' \
		-u $(shell id -u):$(shell id -g) \
		-v $(CURDIR):/local \
			openapitools/openapi-generator-cli:$(OPENAPI_GENERATOR_CLI_VERSION) \
			/bin/sh -c '\
				java -jar /opt/openapi-generator-cli/openapi-generator-cli.jar \
					validate \
					--input-spec /local/api/openapi.yaml'

api-doc-convert:
	@docker run \
		--rm \
		--entrypoint='' \
		-u $(shell id -u):$(shell id -g) \
		-v $(CURDIR):/local \
			openapitools/openapi-generator-cli:$(OPENAPI_GENERATOR_CLI_VERSION) \
			/bin/sh -c '\
				java -jar /opt/openapi-generator-cli/openapi-generator-cli.jar \
					generate \
					--input-spec /local/api/openapi.yaml \
					--generator-name openapi --output /tmp/api/ && \
				cp /tmp/api/openapi.json /local/api'
