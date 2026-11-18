FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# default.pgo di cmd/api otomatis dipakai (-pgo=auto)
RUN CGO_ENABLED=0 go build -trimpath -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -trimpath -o /out/migration ./cmd/migration

FROM gcr.io/distroless/static-debian12:nonroot
# Diisi release.yml dari tag (v1.2.8 -> 1.2.8); kosong = default di config.go.
ARG APP_VERSION
ENV APP_VERSION=${APP_VERSION}
COPY --from=build /out/ /app/
COPY migrations /app/migrations
WORKDIR /app
ENTRYPOINT ["/app/api"]
