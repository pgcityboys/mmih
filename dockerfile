FROM golang:1.22.3-alpine as builder
COPY . /app
WORKDIR /app
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /backapp

FROM gcr.io/distroless/static-debian11 AS runner
COPY --chown=nonroot --from=builder /backapp /app
USER nonroot:nonroot
ENTRYPOINT ["/app"]
