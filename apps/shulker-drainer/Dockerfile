FROM golang:1.19 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /build
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY libs/crds libs/crds
COPY apps/shulker-drainer apps/shulker-drainer

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
  go build -a -o shulker-drainer apps/shulker-drainer/src/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /

COPY --from=builder /build/shulker-drainer .
USER 65532:65532

ENTRYPOINT ["/shulker-drainer"]
