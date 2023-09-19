# syntax=docker/dockerfile:1

FROM ubuntu:22.04

COPY scantool /
COPY scantool.conf /
#COPY run-scantool-docker.sh /
ADD https://github.com/btc-script-explorer/scantool/blob/0.1.0/run-scantool-docker.sh /

COPY web/css /web/css
COPY web/html /web/html
COPY web/js /web/js

CMD ./run-scantool-docker.sh

