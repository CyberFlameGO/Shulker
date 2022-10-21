FROM golang:1.19 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /build
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY packages/crds packages/crds
COPY packages/controllers packages/controllers
COPY packages/resources packages/resources
COPY packages/shulker-operator packages/shulker-operator

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
  go build -a -o shulker-operator packages//shulker-operator/src/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /

COPY --from=builder /build/shulker-operator .
USER 65532:65532

ENTRYPOINT ["/shulker-operator"]
