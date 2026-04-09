FROM golang:1.22-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/rtc-server ./cmd/server

FROM alpine:3.20

RUN adduser -D -H appuser
USER appuser

COPY --from=build /bin/rtc-server /usr/local/bin/rtc-server

EXPOSE 8080

ENTRYPOINT ["rtc-server"]
