#!/usr/bin/env bash
set -euo pipefail
: "${1:?Pass the disposable verification TLS directory}"
mkdir -p "$1"
openssl req -x509 -newkey rsa:2048 -nodes -sha256 -days 2 \
  -keyout "$1/broker.key" -out "$1/broker.crt" \
  -subj '/CN=ucs503-sandbox-broker' \
  -addext 'subjectAltName=DNS:ucs503-sandbox-broker,DNS:localhost,IP:127.0.0.1' \
  -addext 'basicConstraints=critical,CA:TRUE' >/dev/null 2>&1
chmod 600 "$1/broker.key"
chmod 644 "$1/broker.crt"
