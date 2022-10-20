FROM golang:1.19 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /build
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY cmd/shulker-operator/ cmd/shulker-operator/
COPY api/ api/
COPY internal/ internal/

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
  go build -a -o shulker-operator cmd/shulker-operator/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /

COPY --from=builder /shulker-operator .
USER 65532:65532

ENTRYPOINT ["/shulker-operator"]
