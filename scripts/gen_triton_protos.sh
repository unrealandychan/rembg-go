#!/usr/bin/env bash
set -euo pipefail

# Downloads Triton server protos and generates Go bindings into pkg/backends/tritonpb
# Requirements: git, protoc, protoc-gen-go, protoc-gen-go-grpc

OUT_DIR="$(cd "$(dirname "$0")/.." && pwd)/pkg/backends/tritonpb"
TMP_DIR="/tmp/triton-protos-$$"

echo "Generating Triton protos into ${OUT_DIR}"
rm -rf "$TMP_DIR"
git clone --depth 1 https://github.com/triton-inference-server/server.git "$TMP_DIR"

mkdir -p "$OUT_DIR"
PROTO_PATHS=(
  "$TMP_DIR/src/core"
)

PROTO_FILES=(
  "$TMP_DIR/src/core/grpc_service.proto"
  "$TMP_DIR/src/core/infer.proto"
)

protoc \
  --proto_path=${PROTO_PATHS[0]} \
  --go_out="$OUT_DIR" --go_opt=paths=source_relative \
  --go-grpc_out="$OUT_DIR" --go-grpc_opt=paths=source_relative \
  "${PROTO_FILES[@]}"

echo "Generated protos in $OUT_DIR"
