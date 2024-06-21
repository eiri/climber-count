FROM alpine:3.6 as alpine

RUN apk add -U --no-cache ca-certificates

FROM scratch
ENTRYPOINT ["/climber-count"]
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY climber-count climber-count
