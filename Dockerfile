FROM golang:alpine as build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/gophermart ./cmd/gophermart

FROM alpine:latest as run

WORKDIR /app

COPY --from=build /app/gophermart .

EXPOSE 8080
CMD ["./gophermart"]