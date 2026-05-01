FROM golang:1.26.2-alpine AS build

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -trimpath \
    -ldflags='-s -w -extldflags "-static"' \
    -o /out/aviatrix-network-policy-controller \
    .

FROM scratch

COPY --from=build /out/aviatrix-network-policy-controller /aviatrix-network-policy-controller

USER 65532:65532

ENTRYPOINT ["/aviatrix-network-policy-controller"]
