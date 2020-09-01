FROM golang:1.14-alpine
ENV GOOS linux
ENV GOARCH amd64
ENV CGO_ENABLED 0
RUN apk add --no-cache zip
WORKDIR /code
COPY code/go.mod code/go.sum /code/
RUN go mod download
COPY code/* /code/
RUN go build -o main *.go \
    && zip lambda.zip main \
    go test
VOLUME /code