FROM golang:1.16.3-alpine AS build

ENV CGO_ENABLED=0

WORKDIR /src

# Copy Go code to the container
COPY . .
# Download deps usings go mod
RUN go mod download

# dmsg daemon
RUN go build -o /out/dmsgd ./dmsg

FROM ubuntu:20.10

# Copy compiled dmsg daemon
COPY --from=build /out/dmsgd /bin/dmsgd

# Making dmsgd to be an entrypoint
ENTRYPOINT ["/bin/dmsgd", "--log-dir=/dmsgd-data"]