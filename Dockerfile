FROM golang:latest as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV TZ=Asia/Tokyo

WORKDIR /go/src/github.com/reud/unfire
COPY . .
RUN go build main.go

# runtime image
FROM ubuntu:14.04
RUN  apt-get update && apt-get install -y redis-server
RUN /usr/bin/redis-server --daemonize yes
COPY --from=builder /go/src/github.com/reud/unfire /app

CMD /app/main