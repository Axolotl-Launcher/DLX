FROM golang:1.25.0 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY sponsor-gateway ./sponsor-gateway
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/sponsor-gateway ./sponsor-gateway/cmd/server
FROM alpine:3.20
RUN apk add --no-cache wget && adduser -D -H gateway
COPY --from=build /out/sponsor-gateway /sponsor-gateway
# Docker Compose mounts file secrets as root-owned 0400 files. Keep the
# process root only to read these immutable mounts; the image has no shell
# entrypoint or package manager and exposes no host socket.
EXPOSE 8080
ENTRYPOINT ["/sponsor-gateway"]
