FROM --platform=$BUILDPLATFORM golang:1.23 AS builder

WORKDIR /

COPY . .

RUN go mod download

ARG TARGETOS
ARG TARGETARCH

RUN  CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o /output/shifu-akri-adapter *.go

FROM gcr.io/distroless/static-debian12

WORKDIR /
COPY --from=builder /output/shifu-akri-adapter .
COPY templates templates

ENTRYPOINT [ "./shifu-akri-adapter" ]