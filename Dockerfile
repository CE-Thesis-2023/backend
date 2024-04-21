FROM golang:1.21.4-bookworm AS builder

WORKDIR /build
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

RUN apt update && apt install libdlib-dev \
    libblas-dev \
    libatlas-base-dev \
    liblapack-dev \
    libjpeg62-turbo-dev -y

COPY src src

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -a \
    -ldflags "-w -s" \
    -o main src/main.go

FROM debian:bookworm AS runner

COPY models /models

COPY configs.json /configs/configs.json
COPY certs certs/

RUN apt update && apt install ca-certificates \
    libdlib-dev \
    libblas-dev \
    libatlas-base-dev \
    liblapack-dev \
    libjpeg62-turbo-dev -y

WORKDIR /
COPY --from=builder /build/main main

EXPOSE 9000
EXPOSE 9001

ENV CONFIG_FILE_PATH=/configs/configs.json

CMD [ "./main" ]