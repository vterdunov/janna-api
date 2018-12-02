ARG GOLANG_VERSION=1.11.2
ARG OPENAPI_GENERATOR_CLI_VERSION=v3.2.2
# ---------------
FROM golang:${GOLANG_VERSION} AS env

RUN mkdir /user && \
    echo 'janna:x:65534:65534:janna:/:' > /user/passwd && \
    echo 'janna:x:65534:' > /user/group

WORKDIR /gomod
COPY go.mod go.sum ./
RUN go mod download

# ---------------
FROM env AS builder
WORKDIR /build
COPY . .

RUN make compile

# ---------------
FROM builder AS tester

RUN make test

# ---------------
FROM builder AS licence-scanner

ARG GOLICENSE_VERSION=0.1.1
RUN wget https://github.com/mitchellh/golicense/releases/download/v${GOLICENSE_VERSION}/golicense_${GOLICENSE_VERSION}_linux_x86_64.tar.gz && \
		tar xvf golicense_${GOLICENSE_VERSION}_linux_x86_64.tar.gz
RUN ./golicense ./build/ci/golicense.hcl janna-api

# ---------------
FROM openapitools/openapi-generator-cli:${OPENAPI_GENERATOR_CLI_VERSION} AS api
WORKDIR /api
COPY api/openapi.yaml .

RUN java -jar /opt/openapi-generator-cli/openapi-generator-cli.jar \
					validate \
					--input-spec openapi.yaml

# ---------------
FROM api as converter
RUN java -jar /opt/openapi-generator-cli/openapi-generator-cli.jar \
					generate \
					--input-spec openapi.yaml \
					--generator-name openapi --output .

# ---------------
FROM alpine:3.8 AS final

ENV PORT=8080

WORKDIR /janna-api

COPY --from=builder /user/group /user/passwd /etc/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER janna:janna

CMD ["./janna-api"]

COPY --from=converter /api/openapi.json api/openapi.json
COPY --from=builder /build/janna-api janna-api
