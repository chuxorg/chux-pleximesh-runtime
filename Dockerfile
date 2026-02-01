FROM golang:1.24 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /workspace/bin/runtime-daemon ./cmd/runtime-daemon

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /workspace/bin/runtime-daemon /app/runtime-daemon

VOLUME ["/library"]

EXPOSE 8787

ENTRYPOINT ["/app/runtime-daemon"]
