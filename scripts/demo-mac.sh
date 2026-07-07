#!/usr/bin/env bash
# Build tinynginx image and apply mac-demo manifests.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

export SHEPHERD_API="${SHEPHERD_API:-localhost:9876}"
export SHEEP_DATA_DIR="${SHEEP_DATA_DIR:-$ROOT/.run/sheep}"

mkdir -p "$SHEEP_DATA_DIR"

echo "==> SHEEP_DATA_DIR=$SHEEP_DATA_DIR"

if ! ./bin/sheep images 2>/dev/null | grep -q minimal; then
  echo "==> bootstrap minimal image"
  ./bin/sheep bootstrap minimal
fi

echo "==> build + import tinynginx image"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

mkdir -p "$TMP/rootfs"/{bin,etc,dev,proc,sys,tmp}
CGO_ENABLED=0 go build -o "$TMP/rootfs/bin/tinynginx" ./examples/tinynginx/main.go
echo "tinynginx" > "$TMP/rootfs/etc/hostname"
echo "127.0.0.1 localhost" > "$TMP/rootfs/etc/hosts"
( cd "$TMP/rootfs" && tar czf "$TMP/tinynginx.tar.gz" . )
./bin/sheep import tinynginx "$TMP/tinynginx.tar.gz"

DEMO="$ROOT/examples/demo/mac-demo"
for f in \
  deployment-postgres.json \
  deployment-wordpress.json \
  deployment-redis.json \
  deployment-tinynginx.json \
  service-postgres.json \
  service-wordpress.json \
  service-redis.json \
  service-tinynginx.json
do
  echo "==> apply $f"
  ./bin/sheepctl apply -f "$DEMO/$f"
done

echo ""
echo "Done. Check dashboard or:"
echo "  SHEPHERD_API=$SHEPHERD_API ./bin/sheepctl get pods"
echo "  SHEPHERD_API=$SHEPHERD_API ./bin/sheepctl get deployments"
echo "  SHEPHERD_API=$SHEPHERD_API ./bin/sheepctl get services"
