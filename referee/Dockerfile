# Use an official Golang runtime as a parent image
FROM golang:latest

# Set the working directory to /go/src/referee
WORKDIR /go/src/referee

# Copy the source code into the container at /go/src/referee
COPY referee .

COPY common /go/src/common
COPY cowboy /go/src/cowboy

RUN cd /go/src/cowboy && go build

# Build the referee binary
RUN go build -o referee

# Expose port 8080 for the application
EXPOSE 8080

# Define the command to run the application
CMD ["./referee"]
