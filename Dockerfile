FROM golang:1.25.0 AS builder
WORKDIR /go/src/github.com/OwO-Network/DLX
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o deeplx .

FROM alpine:3.20
RUN apk add --no-cache wget
WORKDIR /app
RUN addgroup -S deeplx && adduser -h /app -G deeplx -SH deeplx
USER deeplx:deeplx
COPY --from=builder --chown=deeplx:deeplx /go/src/github.com/OwO-Network/DLX/deeplx /app/deeplx
EXPOSE 1188
ENTRYPOINT ["/app/deeplx"]
