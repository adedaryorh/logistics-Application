FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum ./
COPY pkg/ pkg/
COPY services/ services/
RUN go build -o /bin/server ./services/identity-service/cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/server /server
EXPOSE 8081
CMD ["/server"]
