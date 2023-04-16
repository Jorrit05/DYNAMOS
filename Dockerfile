FROM golang:1.20.3-alpine3.17
ARG NAME
ENV NAME=$NAME

# Set the working directory
WORKDIR /app/cmd/$NAME
RUN mkdir /app/pkg

COPY *.go .

# Copy go.mod and go.sum files into the container
COPY go.mod /app
COPY go.sum /app

# # Copy the /pkg directory
COPY pkg /app/pkg

# Build the application
RUN go build -o /app/$NAME .

EXPOSE 3000

# Start the application
CMD ["/bin/sh", "-c", "/app/$NAME"]