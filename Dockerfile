# syntax=docker/dockerfile:1

FROM ubuntu:22.04

COPY scantool /
COPY web/css /web/css
COPY web/html /web/html
COPY web/js /web/js
COPY VERSION /
COPY docker-run-scantool.sh /

VOLUME config-dir

CMD ./docker-run-scantool.sh

