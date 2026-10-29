#!/usr/bin/env bash
set -euo pipefail

PROTO_DIR=local-dev/proto
GO_MODULE=github.com/ds454/auth
GENERATED_PKG=src/generated

# Make the destination directory
mkdir -p "$GENERATED_PKG"

proto_files=()
m_opts=()

while IFS= read -r -d '' file; do
  rel="${file#"$PROTO_DIR"/}"              # ds454/locomotive/problem/v0/problem.proto
  import_path="$GO_MODULE/$GENERATED_PKG/$(dirname "$rel")"

  proto_files+=("$file")
  m_opts+=("--go_opt=M$rel=$import_path" "--go-grpc_opt=M$rel=$import_path")
done < <(find "$PROTO_DIR" -name '*.proto' -print0)

protoc \
  --proto_path="$PROTO_DIR" \
  --go_out="$GENERATED_PKG" --go_opt=paths=source_relative \
  --go-grpc_out="$GENERATED_PKG" --go-grpc_opt=paths=source_relative \
  "${m_opts[@]}" \
  "${proto_files[@]}"