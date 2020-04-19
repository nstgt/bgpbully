FROM golang:1.13-buster AS builder
WORKDIR /root
RUN apt update
COPY go.mod go.sum /root/
COPY cmd/ /root/cmd/
COPY internal/ /root/internal/
RUN go mod download \
    && cd cmd/bgpdbully \
    && go install

FROM debian:10.3
RUN apt update \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
RUN mkdir -p /bgpdbully
COPY --from=builder /go/bin/bgpdbully /usr/local/bin/