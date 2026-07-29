FROM golang:1.25 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X main.version=feat-eventbus-policy -X main.buildHash=local -X main.buildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    -o bin/controller \
    ./cmd/controller/

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/bin/controller bin/controller
USER 65532:65532
ENTRYPOINT ["./bin/controller"]
