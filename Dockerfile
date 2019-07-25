FROM debian:buster

USER root

RUN mkdir -p /host/proc
RUN mkdir -p /src/linuxmetrics

ENV GOPATH /src/linuxmetrics

ADD . /src/linuxmetrics/
WORKDIR /src/linuxmetrics
RUN /usr/bin/apt-get update
RUN /usr/bin/apt-get install -y git golang make

RUN /usr/bin/go get -d ./...
RUN /usr/bin/make
RUN mv linuxmetrics-logstash /usr/bin
