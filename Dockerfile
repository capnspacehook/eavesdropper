FROM golang:1.19-alpine AS builder

COPY go.mod main.go /build/

WORKDIR /build

ARG CGO_ENABLED=0

RUN go build -ldflags "-w -s" -trimpath -o eavesdropper

FROM alpine

COPY --from=builder /build/eavesdropper /usr/bin/eavesdropper

RUN apk add --no-cache curl nmap

ENTRYPOINT [ "/usr/bin/eavesdropper" ]
