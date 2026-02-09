###############################################################################
#                                BUILD
###############################################################################
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -ldflags="-s -w" -o /app/bin/main main.go

###############################################################################
#                                RUNTIME
###############################################################################
FROM alpine:3.19

RUN apk add --no-cache \
    ttf-dejavu fontconfig curl ca-certificates tzdata \
    && mkdir -p /lib64 \
    && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

ENV WORKDIR=/app
WORKDIR $WORKDIR

COPY --from=builder /app/bin/main $WORKDIR/main
COPY resource                      $WORKDIR/resource
COPY version.txt                   $WORKDIR/version.txt
COPY manifest/i18n                 $WORKDIR/i18n
COPY manifest/fonts/*.ttc          /usr/share/fonts/
COPY manifest/fonts/*.ttf          /usr/share/fonts/

RUN chmod +x $WORKDIR/main

EXPOSE 80

CMD ["./main"]
