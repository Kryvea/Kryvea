FROM golang:1.24.4-bookworm AS go-builder
WORKDIR /kryvea
COPY . .
RUN make

FROM alpine:3.21.2 AS kryvea
# Ensure IANA timezone database is available for time.LoadLocation
RUN apk add --no-cache tzdata
# copy local "locales" directory to the image
WORKDIR /etc/kryvea
COPY ./locales ./locales

WORKDIR /usr/bin
COPY --from=go-builder ./kryvea/build/kryvea .

ENTRYPOINT "/usr/bin/kryvea"
