FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache postgresql-client
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o nob ./nob.go

ARG DB_HOST=postgres
ENV DB_HOST=$DB_HOST

CMD ["./nob", "-clean-s", "notify", "-s", "notify"]
