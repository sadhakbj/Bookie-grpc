# Base stage for building
FROM golang:1.24 AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build stage for server
FROM base AS build-server
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./src/cmd/server/main.go

# Build stage for client
FROM base AS build-client
RUN CGO_ENABLED=0 GOOS=linux go build -o /client ./src/cmd/client/main.go

# Final stage for server
FROM gcr.io/distroless/static-debian12 AS server
WORKDIR /app
COPY --from=build-server /server .
EXPOSE 8020
CMD ["./server"]

# Final stage for client
FROM gcr.io/distroless/static-debian12 AS client
WORKDIR /app
COPY --from=build-client /client .
EXPOSE 8080
CMD ["./client"]
