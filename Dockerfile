FROM golang:alpine as go
WORKDIR /app
ENV GO111MODULE=on

COPY go.mod .
RUN go mod download

COPY . .
RUN go build -o relationship-manager-service ./cmd

FROM alpine

WORKDIR /app

COPY --from=go /app/relationship-manager-service ./relationship-manager-service
COPY run/config.yaml ./config.yaml
CMD ["./relationship-manager-service"]
