FROM golang:1.20.0-alpine3.17 as BUILDER
ADD . /go/src/github.com/requester
WORKDIR /go/src/github.com/requester
RUN cd /go/src/github.com/requester && \
    go build -o requester

FROM alpine:3.17.1
WORKDIR /etc/requester
COPY --from=BUILDER /go/src/github.com/requester/requester /usr/local/bin
RUN apk add --update --no-cache bash jq

CMD ["/usr/local/bin/requester"]