#FROM golang:alpine AS builder
#WORKDIR /build
#RUN apk add --no-cache git && \
#    git clone https://github.com/yotsuya/my_exporter && \
#    cd my_exporter && \
#    go build -v .

FROM openio/sds:16.10-centos-7
WORKDIR /exporter
#COPY --from=builder /build/my_exporter/my_exporter .
COPY ./my_exporter .
CMD /exporter/my_exporter
