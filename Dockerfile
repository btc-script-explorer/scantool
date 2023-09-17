# syntax=docker/dockerfile:1

FROM ubuntu:22.04
ARG scantool_filename

COPY $scantool_filename /$scantool_filename
COPY scantool.conf
COPY web/css /web/css
COPY web/html /web/html
COPY web/js /web/js
COPY docker-run-scantool.sh /

CMD ./docker-run-scantool.sh

