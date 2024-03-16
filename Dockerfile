FROM golang:1.22-bookworm as build
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY sqlc.yaml ./
COPY *.go ./
COPY sql ./sql
COPY internal ./internal

RUN sqlc generate
RUN templ generate
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/stravafy .

FROM debian:bookworm as final
WORKDIR /app
COPY --from=build /app/stravafy /bin/stravafy

RUN update-ca-certificates -v
ENV GIN_MODE=release

EXPOSE 80

CMD ["/bin/stravafy"]