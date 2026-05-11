# syntax=docker/dockerfile:1
ARG BASEIMAGE=gcr.io/distroless/static:nonroot

FROM golang:1.25 AS build
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" \
    -o /out/storage-metrics-apiserver \
    ./cmd/storage-metrics-apiserver

FROM ${BASEIMAGE}
COPY --from=build /out/storage-metrics-apiserver /storage-metrics-apiserver
USER 65534:65534
ENTRYPOINT ["/storage-metrics-apiserver"]
