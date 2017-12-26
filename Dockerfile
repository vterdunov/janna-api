FROM golang:1.9.1-alpine AS build-stage

ARG WORKDIR=/go/src/github.com/vterdunov/janna-api

WORKDIR $WORKDIR

RUN apk add --no-cache git build-base
RUN go get -v github.com/golang/dep/cmd/dep

COPY . $WORKDIR
RUN [ -d 'vendor' ] || make dep
RUN make compile

FROM scratch

ARG PORT=8080
ENV PORT=${PORT}

CMD ["/janna-api"]
COPY --from=build-stage /go/src/github.com/vterdunov/janna-api/janna-api /janna-api
