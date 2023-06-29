FROM golang:1.20 as builder

COPY . /opt/discordant
WORKDIR /opt/discordant

RUN go build ./cmd/discordant

FROM ubuntu:22.04

RUN apt update
RUN apt install -y ca-certificates
RUN apt clean
COPY --from=builder /opt/discordant/discordant /usr/local/bin/discordant
ENTRYPOINT discordant
