FROM golang:1.25-trixie AS build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY qrcode ./qrcode
COPY *.go ./

RUN CGO_ENABLED=1 go build -ldflags "-w -s" -o /qrcode-api

# hadolint ignore=DL3007
FROM gcr.io/distroless/base-debian13:latest AS deploy

# hadolint ignore=DL3045
COPY --from=build /qrcode-api /

EXPOSE 3000
ENTRYPOINT ["/qrcode-api"]
