FROM golang:1.25-alpine AS build

ARG VERSION=dev
ARG VCS_REF=unknown
ARG BUILD_DATE=unknown

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X openclaw/platformapi/internal/buildinfo.Version=${VERSION} -X openclaw/platformapi/internal/buildinfo.Commit=${VCS_REF} -X openclaw/platformapi/internal/buildinfo.BuildDate=${BUILD_DATE}" -o /out/platform-api ./cmd/server \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X openclaw/platformapi/internal/buildinfo.Version=${VERSION} -X openclaw/platformapi/internal/buildinfo.Commit=${VCS_REF} -X openclaw/platformapi/internal/buildinfo.BuildDate=${BUILD_DATE}" -o /out/platform-api-bootstrap ./cmd/bootstrap \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X openclaw/platformapi/internal/buildinfo.Version=${VERSION} -X openclaw/platformapi/internal/buildinfo.Commit=${VCS_REF} -X openclaw/platformapi/internal/buildinfo.BuildDate=${BUILD_DATE}" -o /out/platform-api-migrate ./cmd/migrate \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X openclaw/platformapi/internal/buildinfo.Version=${VERSION} -X openclaw/platformapi/internal/buildinfo.Commit=${VCS_REF} -X openclaw/platformapi/internal/buildinfo.BuildDate=${BUILD_DATE}" -o /out/platform-api-migrationstatus ./cmd/migrationstatus

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata curl \
    && addgroup -S app \
    && adduser -S -G app -u 10001 app

WORKDIR /app
COPY --from=build /out/platform-api /usr/local/bin/platform-api
COPY --from=build /out/platform-api-bootstrap /usr/local/bin/platform-api-bootstrap
COPY --from=build /out/platform-api-migrate /usr/local/bin/platform-api-migrate
COPY --from=build /out/platform-api-migrationstatus /usr/local/bin/platform-api-migrationstatus

USER 10001:10001

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=5s --retries=10 CMD curl -fsS http://localhost:8080/readyz || exit 1

ENTRYPOINT ["platform-api"]


