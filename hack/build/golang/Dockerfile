FROM golang:1.9.2-alpine
RUN echo http://dl-4.alpinelinux.org/alpine/edge/testing >> /etc/apk/repositories
RUN apk update && apk upgrade && apk add --no-cache bash git openssh alpine-sdk libcurl
RUN apk update && apk add cmake pkgconfig && apk add build-base
RUN git clone git://github.com/libgit2/libgit2.git
RUN git config --global user.name "default"
RUN git config --global user.email "default@email.com"
WORKDIR $GOPATH/libgit2
RUN git checkout v0.24.6 && rm -rf build && mkdir build && cd build \
    && cmake .. -DCMAKE_INSTALL_PREFIX=$TARGET -DBUILD_CLAR=OFF && cmake --build . --target install
ENV PKG_CONFIG_PATH /go/libgit2/build:$PKG_CONFIG_PATH
