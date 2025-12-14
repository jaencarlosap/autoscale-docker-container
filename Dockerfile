FROM golang:1.22-alpine AS build
WORKDIR /app
COPY * ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o starter

FROM alpine:3.19
WORKDIR /app
COPY --from=build /app/starter .
EXPOSE 8080
CMD ["./starter"]
