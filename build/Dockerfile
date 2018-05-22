FROM golang:1.10.2-alpine AS build-stage

WORKDIR /go/src/github.com/vterdunov/janna-api

RUN apk add --no-cache make git

COPY . .

RUN make compile

FROM scratch

ARG PORT=8080
ENV PORT=${PORT}

CMD ["/janna-api"]

COPY --from=build-stage /go/src/github.com/vterdunov/janna-api/janna-api /janna-api
