FROM golang:1.15-alpine as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV TZ=Asia/Tokyo

WORKDIR /go/src/github.com/reud/unfire
COPY . .

RUN apk --update add redis
RUN ls
RUN chmod 744 ./tests/start.sh

CMD sh ./tests/start.sh

