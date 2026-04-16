##
## 2 phases docker build
##

#== 1. compile in builder image ====
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ktailor ./cmd/ktailor

#== 2. build runtime image & copy compiled binary into ====
FROM scratch
WORKDIR /
COPY --from=builder /app/ktailor /ktailor
EXPOSE 8443
ENTRYPOINT ["/ktailor"]

