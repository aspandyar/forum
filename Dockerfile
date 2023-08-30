# Use a specific version of the Golang base image for building
FROM golang:1.20-alpine AS build

# Set metadata
LABEL version="1.0" maintainer="asharip <stephen.novel@gmail.com>"

ENV GO111MODULE=on

# Set the working directory
WORKDIR /app

# Copy Go module files for dependency management
COPY go.mod go.sum ./

COPY /ui/. /app/ui

# Download dependencies
RUN go mod download

# Copy the application source code
COPY . .

# Build the application
RUN apk add --no-cache build-base  \
&& go build -o main ./cmd/web/*

# Create a minimal runtime image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the ts.db file and the database initialization script
COPY st.db init-up.sql ./

# Copy the built binary from the build stage
COPY --from=build /app .

# Expose the port your application listens on
EXPOSE 8080

# Define the command to run the application
CMD ["./main"]
