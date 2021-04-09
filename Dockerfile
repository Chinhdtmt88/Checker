FROM golang:1.16.2-alpine AS build

WORKDIR /ZK


ADD . .
RUN go mod download
RUN go get -v ./...
RUN go build -o main .
RUN apk add --no-cache tzdata
ENV TZ Asia/Ho_Chi_Minh
CMD ["go", "run", "main.go"]



