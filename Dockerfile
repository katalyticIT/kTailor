##
## 2 phase docker build
##

#== make image for compiling configurable
ARG BUILDER_IMG=golang
ARG BUILDER_TAG=1.27.4-trixie

#== 1. compile in builder image ====
FROM $BUILDER_IMG:$BUILDER_TAG  AS builder
WORKDIR /app

#== copy go files
COPY go.mod go.sum  .
COPY cmd/           ./cmd/
COPY internal/      ./internal/

#== prepare & compile
RUN go mod tidy     # use 'RUN go get -u golang.org/x/...  &&  go mod tidy' to update central packages
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ktailor ./cmd/ktailor

#== 2. build runtime image & copy compiled binary into ====
FROM scratch
WORKDIR /
COPY --from=builder /app/ktailor /ktailor
EXPOSE 8443
ENTRYPOINT ["/ktailor"]

