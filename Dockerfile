FROM golang:1.21.4-alpine3.18 AS builder

WORKDIR /build
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY src src

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a \
    -ldflags "-w -s" \
    -o main src/main.go

FROM scratch AS runner

ENV CONFIG_FILE_PATH=configs.json

COPY configs.json configs.json
COPY src/templates templates
COPY --from=builder /build/main main

CMD [ "./main" ]