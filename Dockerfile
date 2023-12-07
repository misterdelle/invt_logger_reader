FROM golang:1.21-alpine3.17 as builder

WORKDIR /build

COPY . .

RUN go mod download && CGO_ENABLED=0

RUN go build -ldflags "-s -w" -o invt-logger-reader

FROM alpine:3.16.4

WORKDIR /

RUN apk upgrade --no-cache --ignore alpine-baselayout --available && \
    apk --no-cache add ca-certificates tzdata && \
    rm -rf /var/cache/apk/*

COPY --from=builder /build/invt-logger-reader .
RUN chmod +x invt-logger-reader

ENTRYPOINT ["/invt-logger-reader"]
