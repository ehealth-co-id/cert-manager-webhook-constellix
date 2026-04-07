# syntax=docker/dockerfile:1

ARG GO_VERSION=1.23.4

FROM golang:${GO_VERSION}-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev
ARG REVISION=

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath \
    -ldflags="-s -w" \
    -o /out/webhook .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build --chmod=0555 /out/webhook /webhook

ARG VERSION=dev
ARG REVISION=
LABEL org.opencontainers.image.title="cert-manager-webhook-constellix" \
      org.opencontainers.image.description="ACME DNS01 webhook for Constellix" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}"

USER nonroot:nonroot
ENTRYPOINT ["/webhook"]
