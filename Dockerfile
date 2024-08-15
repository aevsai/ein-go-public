# Base image
FROM golang:1.19

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

# Download and install the project dependencies
RUN go mod download

# Copy the rest of the project files to the working directory
COPY . .

# Build the project
RUN go build -o main .

# Set the command to run the compiled binary
CMD ["./main"]