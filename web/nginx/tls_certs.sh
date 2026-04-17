#!/bin/sh

pem_file="/etc/nginx/ssl/kryvea.local.pem"
key_file="/etc/nginx/ssl/kryvea.local.key"

if [ ! -f "$key_file" ] || [ ! -f "$pem_file" ]; then
    mkdir -p "/etc/nginx/ssl"
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout "$key_file" -out "$pem_file" \
        -subj "/C=IT/ST=/L=/O=kryvea/CN=kryvea.local"
    chmod 600 "$pem_file"
    chmod 600 "$key_file"
fi