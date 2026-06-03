# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:alpine AS builder

RUN apk add --no-cache upx ca-certificates

WORKDIR /app

COPY go.mod .
COPY main.go .

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o weather-app .

RUN upx --best --lzma weather-app || true

FROM scratch

LABEL org.opencontainers.image.authors="Kacper Kantarowicz"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/weather-app /weather-app

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/weather-app", "-health"]

CMD ["/weather-app"]