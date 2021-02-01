FROM golang:latest as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV TZ=Asia/Tokyo

WORKDIR /go/src/github.com/reud/unfire
COPY . .
RUN go build main.go

# runtime image
FROM ubuntu:20.04
RUN  apt-get update && apt-get install -y redis-server
RUN apt-get install -y ca-certificates
RUN /usr/bin/redis-server --daemonize yes
COPY --from=builder /go/src/github.com/reud/unfire /app
ADD start.sh /
RUN chmod 744 /start.sh

CMD /start.sh