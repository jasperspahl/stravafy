FROM golang:1.22-bookworm as build
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN apt update && apt install -y ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY sqlc.yaml ./
COPY *.go ./
COPY sql ./sql
COPY internal ./internal

RUN sqlc generate
RUN templ generate
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a --installsuffix cgo -v -tags netgo -ldflags '-extldflags "-static"' -o /app/stravafy .

FROM scratch as final
WORKDIR /app

COPY --from=build \
    /etc/ssl/certs/ca-certificates.crt \
    /etc/ssl/certs/ca-certificates.crt

COPY --from=build /app/stravafy /bin/stravafy

ENV GIN_MODE=release

EXPOSE 80

CMD ["/bin/stravafy"]