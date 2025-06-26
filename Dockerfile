FROM golang:1.24.4-alpine as builder

RUN mkdir -p "/vhs/"
COPY . /vhs/

WORKDIR /vhs/
RUN go mod download

RUN apk update \
    && apk upgrade \
    && apk add ffmpeg

RUN go build -a -installsuffix cgo -o ./vhs cmd/app/main.go

EXPOSE 8080

CMD ["./vhs", "serve", "--http=0.0.0.0:8080"]