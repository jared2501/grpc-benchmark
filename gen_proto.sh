#!/usr/bin/env bash
set -euo pipefail
set -x
cd "$(dirname "$0")"
protoc -I=. --go_out=plugins=grpc,paths=source_relative:. *.proto