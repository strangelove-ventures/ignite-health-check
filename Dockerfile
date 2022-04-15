FROM golang:alpine AS build-env
WORKDIR /build
ADD . .
RUN apk add --update git
RUN go build

FROM alpine:edge
COPY --from=build-env /build/ignite-health-check /usr/local/bin/ihc
EXPOSE 1251
CMD ihc
