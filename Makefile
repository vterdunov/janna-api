PROG_NAME = janna-api
IMAGE_NAME = vterdunov/$(PROG_NAME)

PORT ?= 8080
COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%S')
PROJECT ?= github.com/vterdunov/${PROG_NAME}

GO_VARS=CGO_ENABLED=0 GOOS=linux GOARCH=amd64
GO_LDFLAGS :="
GO_LDFLAGS += -s -w
GO_LDFLAGS += -X ${PROJECT}/pkg/version.Commit=${COMMIT}
GO_LDFLAGS += -X ${PROJECT}/pkg/version.BuildTime=${BUILD_TIME}
GO_LDFLAGS +="


all: check test docker

.PHONY: docker
docker:
	docker build --tag=$(IMAGE_NAME):$(COMMIT) --tag=$(IMAGE_NAME):latest .

.PHONY: push
push:
	docker push $(IMAGE_NAME):$(COMMIT)

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
	go run ./cmd/janna/server.go

.PHONY: test
test:
	go test -v ./...

.PHONY: check
check:
	@gometalinter ./...

.PHONY: clean
clean:
	@rm -f ${PROG_NAME}
