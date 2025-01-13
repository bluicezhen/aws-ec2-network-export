FROM public.ecr.aws/docker/library/golang:1.23.4-alpine3.21 AS builder

WORKDIR /app

COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o network-export

FROM public.ecr.aws/docker/library/alpine:3.21.0

# Install required system packages for ethtool
# RUN apk add --no-cache ethtool

EXPOSE 2112

# Copy binary from builder
COPY --from=builder /app/network-export /network-export

ENTRYPOINT ["/network-export"]
