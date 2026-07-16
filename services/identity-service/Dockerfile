FROM golang:1.25.6-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum ./
COPY pkg/ pkg/
COPY services/ services/
COPY tests/go.mod tests/go.mod
RUN go build -o /bin/server ./services/identity-service/cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/server /server
EXPOSE 8081
CMD ["/server"]
