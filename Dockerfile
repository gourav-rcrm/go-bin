FROM golang:1-alpine as build

WORKDIR /app
COPY . .
RUN go build main.go

FROM alpine:latest

WORKDIR /app
COPY --from=build /app /app

EXPOSE 8180
ENTRYPOINT ["./main"]
