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

GOLANGCI_LINTER_VERSION = v1.12.3
OPENAPI_GENERATOR_CLI_VERSION = v3.2.2

all: lint docker

.PHONY: help
help: ## Display this help message
	@cat $(MAKEFILE_LIST) | grep -e "^[-a-zA-Z_\.]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: docker
docker: ## Build Docker container
	docker build --tag=$(IMAGE_NAME):$(COMMIT) --tag=$(IMAGE_NAME):latest --build-arg=GITHUB_TOKEN=${GITHUB_TOKEN} --file build/Dockerfile .

.PHONY: push
push: ## Push docker container to registry
	docker tag $(IMAGE_NAME):$(COMMIT) $(IMAGE_NAME):$(TAG)
	docker push $(IMAGE_NAME):$(TAG)

.PHONY: compile
compile: clean ## Compile Janna
	$(GO_VARS) go build -v -ldflags $(GO_LDFLAGS) -o $(PROG_NAME) ./cmd/janna/server.go

.PHONY: cgo-compile
cgo-compile: clean
	go build -v -o $(PROG_NAME) ./cmd/janna/server.go

.PHONY: run
run: ## Extract env variables from .env and run Janna with race detector
	@env `cat .env | grep -v ^# | xargs` go run -race ./cmd/janna/server.go

compile-and-run: compile ## Extract env variables from .env. Compile and run Janna
	@env `cat .env | grep -v ^# | xargs` ./janna-api

.PHONY: dc
dc: dc-clean ## Run project using docker-compose. Autobuild when files was changed.
	docker-compose -f deploy/docker-compose.dev.yml up --build

dc-clean:
	docker-compose -f deploy/docker-compose.dev.yml down --volumes

.PHONY: test
test: ## Run tests
	go test -v -race ./...

.PHONY: lint
lint: ## Run linters
	@echo Linting...
	@docker run --tty --rm -v $(CURDIR):/lint -w /lint golangci/golangci-lint:$(GOLANGCI_LINTER_VERSION) golangci-lint run

.PHONY: clean
clean: ## Removes binary file
	@rm -f ${PROG_NAME}

api-doc-validate: ## Validate OpenAPI spec
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

api-doc-convert: ## Convert OpenAPI spec from yaml to json format
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

.PHONY: proto
proto: ## Generate protocol buffers code
	@ocker run -it --rm -v ${PWD}:/work uber/prototool prototool generate
