FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/climber-count ./...

FROM alpine:3.22 AS alpine
RUN apk add -U --no-cache ca-certificates

FROM scratch
ENTRYPOINT ["/climber-count"]
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/climber-count /climber-count
