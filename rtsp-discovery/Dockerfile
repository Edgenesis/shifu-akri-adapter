FROM --platform=$BUILDPLATFORM golang:1.23 AS builder

WORKDIR /discovery

COPY . .

RUN go mod download

ARG TARGETOS
ARG TARGETARCH

RUN  CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o /output/discovery *.go

FROM gcr.io/distroless/static-debian12

WORKDIR /
COPY --from=builder /output/discovery .

ENTRYPOINT [ "./discovery" ]