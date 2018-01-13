PROG_NAME = janna-api

PORT?=8080
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
PROJECT?=github.com/vterdunov/${PROG_NAME}

GO_VARS=CGO_ENABLED=0 GOOS=linux GOARCH=amd64
GO_LDFLAGS=-v -ldflags="-s -w \
		-X ${PROJECT}/version.Release=${RELEASE} \
    -X ${PROJECT}/version.Commit=${COMMIT} \
		-X ${PROJECT}/version.BuildTime=${BUILD_TIME}"

all: dep test compile api-doc

container:
	docker build -t $(PROG_NAME) .

container-run:
	docker run -it --rm --name=$(PROG_NAME) $(PROG_NAME)

dep:
	@dep ensure -v

compile: clean
	 $(GO_VARS) go build $(GO_LDFLAGS) -o $(PROG_NAME)

start: compile
	PORT=${PORT} ./${PROG_NAME}

test:
	go test -v ./...

clean:
	rm -f ${PROG_NAME}

api-doc:
	swagger generate spec --scan-models --output=swagger.json
	swagger validate swagger.json

serve-api-doc:
	swagger serve swagger.json
