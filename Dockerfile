###################################
#Build stage
FROM golang:1.12-alpine AS build-env

#Build deps
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk --no-cache add git
COPY ./ /gopher
WORKDIR /gopher
RUN go build -tags 'bindata'

FROM alpine:3.9

EXPOSE 8000

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk --no-cache add \
    bash \
    ca-certificates \
    curl \
    gettext \
    s6 \
    su-exec \
    tzdata

VOLUME ["/data"]

ENTRYPOINT ["/app/gopher/gopher"]

COPY --from=build-env ./gopher/gopher /app/gopher/gopher