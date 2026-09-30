FROM golang:1.25-bookworm AS go-builder
WORKDIR /kryvea
COPY . .
RUN make

FROM alpine:3.21.2 AS kryvea-migrate
RUN apk add --no-cache ca-certificates
COPY --from=go-builder /kryvea/build/kryvea-migrate /usr/local/bin/kryvea-migrate
ENTRYPOINT ["/usr/local/bin/kryvea-migrate"]

FROM alpine:3.21.2 AS kryvea
# Ensure IANA timezone database is available for time.LoadLocation
RUN apk add --no-cache tzdata
WORKDIR /etc/kryvea
COPY ./locales ./locales

WORKDIR /usr/bin
COPY --from=go-builder /kryvea/build/kryvea .
COPY --from=go-builder /kryvea/build/kryvea-migrate .

ENTRYPOINT ["/usr/bin/kryvea"]
