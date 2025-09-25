# Golang backend environment build container
FROM golang:1.23-alpine AS build

WORKDIR /build

ADD . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o bin/kloud-dns .

# Deploy container
FROM gcr.io/distroless/static:nonroot

COPY --from=build --chown=nonroot:nonroot /build/bin/kloud-dns /kloud-dns

#Start The Project
ENTRYPOINT ["/kloud-dns"]
