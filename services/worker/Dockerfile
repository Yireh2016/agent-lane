# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
# Reason: the optional proxy_ca secret lets builds run behind a TLS-intercepting proxy
# (see docker-compose.proxy.yml); without it the system roots are used.
RUN --mount=type=secret,id=proxy_ca,required=false \
    if [ -f /run/secrets/proxy_ca ]; then export SSL_CERT_FILE=/run/secrets/proxy_ca; fi; \
    go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/worker ./cmd/worker && CGO_ENABLED=0 go build -o /out/gencerts ./cmd/gencerts

FROM alpine:3.20
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
RUN adduser -D -u 10001 app
COPY --from=build /out/ /usr/local/bin/
USER app
ENTRYPOINT ["/usr/local/bin/worker"]
