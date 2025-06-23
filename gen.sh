#!/bin/bash

set -e

BUF_MODULE="buf.build/project-kessel/inventory-api"
PROTO_OUT_DIR="_proto"

echo "Creating proto directories..."
mkdir -p "${PROTO_OUT_DIR}"

echo "Exporting protos from Buf.build..."
buf export "${BUF_MODULE}" --output "${PROTO_OUT_DIR}"