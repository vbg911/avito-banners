# Сборка
FROM golang:latest AS build

WORKDIR /go/src/avito-banner
COPY go.mod .
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o avito-banner ./cmd/banners-api

# Запуск
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=build /go/src/avito-banner/avito-banner .

CMD ["./avito-banner"]
