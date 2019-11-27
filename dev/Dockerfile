FROM openio/sds:18.04
WORKDIR /work
ARG VERSION=1.13.4
ARG OS=linux
ARG ARCH=amd64
RUN curl -SL https://dl.google.com/go/go$VERSION.$OS-$ARCH.tar.gz | tar -xzC /usr/local
ENV PATH=$PATH:/usr/local/go/bin
