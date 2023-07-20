FROM golang:1.18-alpine as builder

# Provide environment variable when build this Dockerfile
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    TZ=Asia/Jakarta \
    GIT_TERMINAL_PROMPT=1

# First Update
RUN apk update  \
    && apk upgrade \
    && apk add --no-cache \
    && apk add git \
    # ca-certificates \
    # && update-ca-certificates 2>/dev/null || true \
    && apk add --no-cache tzdata  \
    && cp /usr/share/zoneinfo/${TZ} /etc/localtime  \
    && echo $TZ >  /etc/timezone  \
    && mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

WORKDIR /app/
COPY . /app/
RUN go build -o /app/main ./cmd/app/

FROM alpine:3
WORKDIR /app/
COPY --from=builder /app/main /app/
CMD /app/main