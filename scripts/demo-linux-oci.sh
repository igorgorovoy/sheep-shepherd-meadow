#!/usr/bin/env bash
# Pull OCI images and apply linux-oci demo (Linux + cgroups v2 required).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

export SHEPHERD_API="${SHEPHERD_API:-localhost:9876}"
export SHEEP_DATA_DIR="${SHEEP_DATA_DIR:-/var/lib/sheep}"

if [[ "$(uname -s)" != "Linux" ]]; then
  echo "linux-oci demo requires Linux. Use ./scripts/demo-mac.sh on macOS." >&2
  exit 1
fi

echo "==> pull images (requires network)"
sudo -E ./bin/sheep pull postgres:16-alpine
sudo -E ./bin/sheep pull wordpress:6-apache
sudo -E ./bin/sheep pull nginx:alpine

DEMO="$ROOT/examples/demo/linux-oci"
for f in \
  deployment-postgres.json \
  deployment-wordpress.json \
  deployment-nginx.json \
  service-postgres.json \
  service-wordpress.json \
  service-nginx.json
do
  echo "==> apply $f"
  ./bin/sheepctl apply -f "$DEMO/$f"
done

echo "Applied. Wait for replication controller, then: sheepctl get pods"
