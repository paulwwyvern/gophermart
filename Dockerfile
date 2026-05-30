FROM golang:alpine as build

WORKDIR /app

COPY go.mod .
COPY go.sum .

# Download the Go module dependencies
RUN go mod download

COPY . .

RUN go build -o /app/gophermart ./cmd/gophermart

FROM alpine:latest as run

WORKDIR /app

# Copy the application executable from the build image
COPY --from=build /app/gophermart .

EXPOSE 8080
CMD ["./gophermart"]