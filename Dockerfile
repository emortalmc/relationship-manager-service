FROM --platform=$BUILDPLATFORM golang:1.20 AS build

WORKDIR /app

## Copy go.mod and go.sum files, download dependencies so they are cached
COPY go.mod go.sum ./
RUN go mod download

# Copy sources
COPY cmd ./cmd
COPY internal ./internal

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-s -w" -o relationship-manager-service ./cmd

FROM --platform=$BUILDPLATFORM alpine:3.14 AS app

WORKDIR /app

COPY --from=build /app/relationship-manager-service ./relationship-manager-service
COPY run/config.yaml ./config.yaml
CMD ["./relationship-manager-service"]
