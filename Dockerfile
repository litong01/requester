FROM golang:1.22.2-alpine3.19 as BUILDER
ADD . /go/src/github.com/requester
WORKDIR /go/src/github.com/requester
RUN cd /go/src/github.com/requester && \
    go build -o requester

FROM alpine:3.19
# Install necessary software
RUN apk add --update --no-cache bash jq doas supercronic shadow

# Add non root user
RUN adduser -D -h /home/requester requester && \
    mkdir -p /home/requester/data && chown requester /home/requester/data && \
    adduser requester wheel && \
    echo "permit persist :wheel" > /etc/doas.d/doas.conf

COPY --from=BUILDER /go/src/github.com/requester/requester /usr/local/bin
COPY ./config.yaml /home/requester/config.yaml 
COPY ./crontab /home/requester/crontab

# Setup the crontab

USER requester
ENV DATAROOTDIR=/home/requester/data
ENV CONFIG=/home/requester/config.yaml
WORKDIR /home/requester

ENTRYPOINT ["/usr/bin/supercronic", "/home/requester/crontab"]
