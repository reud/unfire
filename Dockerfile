FROM golang:latest as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
WORKDIR /go/src/github.com/reud/unfire
COPY . .
RUN go build main.go

# runtime image
FROM alpine
COPY --from=builder /go/src/github.com/reud/unfire /app

CMD /app/main