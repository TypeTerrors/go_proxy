FROM golang:1.24-alpine AS stage1
WORKDIR /app
COPY . .
RUN go mod tidy
WORKDIR /app/cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-X main.Version=${VERSION}" -o main /app/main . 

FROM alpine:latest
WORKDIR /app
COPY --from=stage1 /app/main /app/main
EXPOSE 80
CMD ["./main"]

