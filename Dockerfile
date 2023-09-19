# syntax=docker/dockerfile:1

FROM ubuntu:22.04

COPY scantool /
COPY scantool.conf /
COPY VERSION /
COPY run-scantool-docker.sh /

COPY web/css /web/css
COPY web/html /web/html
COPY web/js /web/js

CMD ./run-scantool-docker.sh

