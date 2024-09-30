#!/bin/bash

CERTS_DIR="./ssl-certs"

# Paths for certificate files
PRIVATE_KEY="$CERTS_DIR/server.key"
CSR="$CERTS_DIR/server.csr"
CERT="$CERTS_DIR/server.crt"

# Create the certificates directory if it doesn't exist
mkdir -p "$CERTS_DIR"

# Generate a private key
openssl genrsa -out "$PRIVATE_KEY" 2048

# Create a certificate signing request (CSR)
openssl req -new -key "$PRIVATE_KEY" -out "$CSR" -subj "/C=US/ST=State/L=City/O=Organization/OU=Unit/CN=localhost"

# Sign the CSR to create the certificate
openssl x509 -req -days 365 -in "$CSR" -signkey "$PRIVATE_KEY" -out "$CERT"

# Secure the private key
chmod 600 "$PRIVATE_KEY"

# Remove the CSR as it's no longer needed
rm "$CSR"
