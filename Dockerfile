FROM golang:1.15rc1-alpine

# https://github.com/ishaanbahal/golang-alpine-vips/blob/master/Dockerfile
ARG VIPS_VERSION="8.9.2"
RUN wget https://github.com/libvips/libvips/releases/download/v${VIPS_VERSION}/vips-${VIPS_VERSION}.tar.gz
RUN apk update && apk add automake build-base pkgconfig glib-dev gobject-introspection libxml2-dev expat-dev jpeg-dev libwebp-dev libpng-dev gcc g++ make
# Exit 0 added because warnings tend to exit the build at a non-zero status
RUN tar -xf vips-${VIPS_VERSION}.tar.gz && cd vips-${VIPS_VERSION} && ./configure && make && make install && ldconfig; exit 0
RUN apk del automake build-base

COPY . /app
WORKDIR /app
RUN go get
RUN go build
CMD ./imgate
