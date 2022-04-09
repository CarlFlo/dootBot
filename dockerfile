### Multi-stage build ###

## First stage (building)

FROM golang:latest AS builder

WORKDIR /app

# Copy everything
COPY . .

# Download requirements
RUN go mod download

# Builds the executable
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .


## Second stage (running)

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app .

# Changes the user to non-root. Reduces attack surface
USER 1000

# Runs the app
CMD ["./app"]  

