FROM golang:1.22-alpine as build
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
RUN CGO_ENABLED=0 GOOS=linux go build -o /stravafy .

FROM alpine:latest as final
WORKDIR /app
COPY --from=build /app/stravafy /bin/stravafy

ENV GIN_MODE=release

EXPOSE 80

CMD ["/bin/stravafy"]