FROM golang:1.22.2-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

ARG TARGETOS
ARG TARGETARCH

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o /out/app .

FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /out/app /app

USER nonroot:nonroot

ENTRYPOINT ["/app"]