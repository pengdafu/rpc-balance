FROM golang:1.15.1 as builder

ENV GOPROXY="https://goproxy.cn,direct"
ENV GO111MODULE=on

WORKDIR /build/go/

COPY . .

RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-s -w" -installsuffix cgo -o app cmd/main.go

FROM alpine:latest as prod

WORKDIR /go/

COPY --from=builder /build/go/app .

CMD ["./app"]