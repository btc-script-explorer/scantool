# syntax=docker/dockerfile:1

FROM ubuntu:22.04

COPY scantool /scantool
COPY scantool.conf /scantool.conf
COPY web/css /web/css
COPY web/html /web/html
COPY web/js /web/js
COPY docker-run-scantool.sh /

CMD ./docker-run-scantool.sh

