FROM golang:1.19-alpine as base
ENV GOOS linux
ENV GOARCH amd64
ENV CGO_ENABLED 0
RUN apk add --no-cache zip
WORKDIR /code
COPY code/go.mod code/go.sum /code/
RUN go mod download
COPY code/* /code/

FROM base as build
RUN go build -o main *.go \
    && zip lambda.zip main

FROM base as test
RUN go test
RUN wget -O- -nv 'https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh' \
    | sh -s -- -b "$(go env GOPATH)/bin" 'v1.30.0'
RUN golangci-lint run .

FROM alpine as release
WORKDIR /code
COPY --from=build /code/lambda.zip /code/
VOLUME /code