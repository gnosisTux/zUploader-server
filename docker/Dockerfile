FROM golang:1.25-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
RUN git clone https://github.com/gnosisTux/zUploader-server.git .
RUN go mod download
RUN go build -o zuploader main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/zuploader .
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/rc ./rc
RUN mkdir -p /srv/zuploader/uploads
EXPOSE 8002
CMD ["./zuploader"]
