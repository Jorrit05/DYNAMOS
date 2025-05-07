FROM golang:1.22.1-bullseye AS builder
# FROM test:latest AS builder
ARG NAME
ENV NAME=$NAME

# Set the working directory
WORKDIR /app/cmd/$NAME
RUN mkdir /app/pkg
RUN mkdir /app/etcd

# RUN mkdir /res

COPY *.go ./

# COPY ./synthetic_data_sample.csv /res

# Copy go.mod and go.sum files into the container
COPY go.mod /app
COPY go.sum /app

# # Copy the /pkg directory
COPY pkg /app/pkg

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/$NAME .

# 2. Running the Application in a Small Image
FROM alpine:3.18.2
ARG NAME
ENV NAME=$NAME
RUN apk --no-cache add ca-certificates

RUN mkdir /res
COPY synthetic_data_sample.csv /res
# COPY synthetic_sdv.csv /res

WORKDIR /app/

# Copy the pre-built binary from the previous stage
COPY --from=builder /go/bin/$NAME .

# Start the application
CMD ["/bin/sh", "-c", "/app/$NAME || ls -al /app"]
