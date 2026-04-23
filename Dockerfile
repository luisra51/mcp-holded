# Build stage
FROM golang:1.25.7 AS build

WORKDIR /src
ARG TARGETOS=linux
ARG TARGETARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/mcp-holded ./cmd/mcp-holded

# Runtime stage
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=build /out/mcp-holded /app/mcp-holded

ENV HOLDED_API_BASE=https://api.holded.com/api/invoicing/v1

ENTRYPOINT ["/app/mcp-holded"]
