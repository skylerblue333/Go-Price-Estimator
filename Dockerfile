FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sky-price ./main.go

FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app -u 10001 app
WORKDIR /app
COPY --from=builder /out/sky-price /app/sky-price
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/app/sky-price"]
