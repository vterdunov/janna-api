PROG_NAME = janna-api

PORT ?= 8080
COMMIT ?= $(shell git rev-parse --short HEAD)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%S')
PROJECT ?= github.com/vterdunov/${PROG_NAME}

GO_VARS=CGO_ENABLED=0 GOOS=linux GOARCH=amd64
GO_LDFLAGS :="
GO_LDFLAGS += -s -w
GO_LDFLAGS += -X ${PROJECT}/version.Commit=${COMMIT}
GO_LDFLAGS += -X ${PROJECT}/version.BuildTime=${BUILD_TIME}
GO_LDFLAGS +="

all: dep check test compile api-doc

.PHONY: container
container:
	docker build -t $(PROG_NAME) .

.PHONY: container-run
container-run:
	docker run -it --rm --name=$(PROG_NAME) $(PROG_NAME)

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
start: cgo-compile
	@./${PROG_NAME}

.PHONY: test
test:
	go test -v ./...

.PHONY: check
check:
	@gometalinter \
		--vendor \
		--disable-all \
		--enable=vet \
		--enable=vetshadow \
		--enable=golint \
		--enable=ineffassign \
		--enable=goconst \
		--enable=deadcode \
		--enable=gofmt \
		--enable=misspell \
		--enable=dupl \
		--enable=gotype \
		--enable=goimports \
		--enable=gotypex \
		--tests \
		--aggregate \
		./...

.PHONY: clean
clean:
	@rm -f ${PROG_NAME}

.PHONY: api-doc
api-doc:
	swagger generate spec --scan-models --output=swagger.json
	swagger validate swagger.json

.PHONY: serve-api-doc
serve-api-doc:
	swagger serve swagger.json
