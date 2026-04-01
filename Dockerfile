FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /vergo ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /vergo /vergo

EXPOSE 8080

ENTRYPOINT ["/vergo"]
