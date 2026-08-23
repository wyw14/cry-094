# syntax=docker/dockerfile:1.7
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /out/scriptscope ./cmd/server

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/scriptscope /usr/local/bin/scriptscope
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/scriptscope"]
