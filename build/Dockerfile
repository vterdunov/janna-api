FROM golang:1.11.2-alpine3.8 AS builder

RUN mkdir /user && \
    echo 'janna:x:65534:65534:janna:/:' > /user/passwd && \
    echo 'janna:x:65534:' > /user/group

WORKDIR /build

RUN apk add --no-cache make git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make compile

FROM alpine:3.8

ENV PORT=8080

WORKDIR /janna-api

COPY --from=builder /user/group /user/passwd /etc/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER janna:janna

CMD ["./janna-api"]

COPY api/openapi.json api/openapi.json
COPY --from=builder /build/janna-api janna-api
