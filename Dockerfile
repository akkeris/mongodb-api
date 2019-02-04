#
#
# Intial FROM is a fix for Jenkins docker plugin
# not being able to parse the 'as xxxxx' in the FROM command
#
FROM golang:1.9-alpine3.7
FROM golang:1.9-alpine3.7 as builder

RUN apk update; \
    apk add git openssl ca-certificates

RUN go get -u github.com/golang/dep/cmd/dep

ENV APPNAME mongodb-api

ARG PKGS="mongodb-api/db mongodb-api/server"

ARG VAULT_ADDR
ENV VAULT_ADDR=${VAULT_ADDR}

ARG VAULT_TOKEN
ENV VAULT_TOKEN=${VAULT_TOKEN}

ARG MONGODB_SECRET
ENV MONGODB_SECRET=${MONGODB_SECRET}

ARG MONGODB_API_RUNTIME=development
ENV MONGODB_API_RUNTIME=${MONGODB_API_RUNTIME}

WORKDIR /go/src/${APPNAME}

COPY . .

RUN dep ensure

RUN go test -v ${PKGS}

RUN go build -o ${APPNAME}

#
#
#
FROM alpine:3.7

RUN apk update; \
    apk add openssl ca-certificates

ENV APPNAME mongodb-api

ENV MONGODB_API_RUNTIME=development

ARG PORT=4040
ENV PORT=${PORT}

WORKDIR /app

COPY --from=builder /go/src/${APPNAME}/${APPNAME} .

CMD /app/${APPNAME}
