FROM golang:1.10.2-alpine AS build-stage

WORKDIR /go/src/github.com/vterdunov/janna-api

RUN apk add --no-cache make git

COPY . .

RUN make compile

FROM scratch

ENV PORT=8080

WORKDIR /janna-api

CMD ["./janna-api"]

COPY api/openapi.json api/openapi.json
COPY --from=build-stage /go/src/github.com/vterdunov/janna-api/janna-api janna-api
