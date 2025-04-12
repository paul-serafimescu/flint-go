FROM golang:1.24-alpine as builder
WORKDIR /app
COPY ../ndnd /app
RUN apk add --no-cache make git && \
    make && \
    make install

FROM alpine:latest
COPY --from=builder /usr/local/bin/ndnd /usr/local/bin/ndnd
COPY ../configs/fw.config.yml /etc/ndn/fw.config.yml
EXPOSE 6363
ENTRYPOINT ["ndnd", "fw", "run", "/etc/ndn/fw.config.yml"]
