# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine3.24 AS build
WORKDIR /src

ARG TARGETOS
ARG TARGETARCH

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/go-api .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/go-api /go-api

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/go-api"]
