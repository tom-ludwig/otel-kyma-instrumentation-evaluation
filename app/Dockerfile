FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o demo-app .

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/demo-app /demo-app
EXPOSE 8080
ENTRYPOINT ["/demo-app"]
