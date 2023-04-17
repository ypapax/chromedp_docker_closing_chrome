# Compile app binary
FROM golang:latest as build-env

ENV GO111MODULE=on

WORKDIR /go/src
COPY main.go main.go
COPY mem_stat.go mem_stat.go
COPY selenium.go selenium.go
COPY go.mod go.mod
COPY go.sum go.sum

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -installsuffix cgo -o /app

FROM arunvelsriram/utils
USER root
RUN apt upgrade -y
RUN apt-get install curl gnupg -y
RUN curl -fsSL https://github.com/rabbitmq/signing-keys/releases/download/2.0/rabbitmq-release-signing-key.asc | apt-key add -
RUN apt-key adv --keyserver "keyserver.ubuntu.com" --recv-keys "F77F1EDA57EBB1CC"
RUN apt-get install apt-transport-https
RUN echo "deb http://ppa.launchpad.net/rabbitmq/rabbitmq-erlang/ubuntu bionic main" > /etc/apt/sources.list.d/bintray.rabbitmq.list
RUN echo "deb-src http://ppa.launchpad.net/rabbitmq/rabbitmq-erlang/ubuntu bionic main" >> /etc/apt/sources.list.d/bintray.rabbitmq.list
RUN apt-get update
RUN apt install lsof

RUN DEBIAN_FRONTEND="noninteractive" apt-get -y install tzdata

RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
RUN dpkg -i google-chrome-stable_current_amd64.deb; apt-get -fy install

COPY --from=build-env /app /app


CMD ["/app"]
