FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/ingestor ./cmd/ingestor

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/ingestor /ingestor
EXPOSE 8081
USER nonroot:nonroot
ENTRYPOINT ["/ingestor"]
