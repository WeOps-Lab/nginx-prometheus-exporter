#
# build container
#
FROM --platform=linux/amd64 golang:1.19-alpine as builder
WORKDIR /

COPY nginx_exporter /nginx_exporter

EXPOSE     9113
ENTRYPOINT [ "/nginx_exporter" ]
