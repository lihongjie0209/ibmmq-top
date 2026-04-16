#!/usr/bin/env bash
# build.sh - Build mq-top binary inside Linux Docker container
# Output: ./out/mq_top  (Linux amd64 binary)

set -e

IMAGE="mq-top-builder"
OUTDIR="$(dirname "$0")/out"
mkdir -p "$OUTDIR"

echo "[1/2] Building Docker image..."
docker build -t "$IMAGE" -f Dockerfile.build .

echo "[2/2] Compiling mq_top binary..."
docker run --rm -v "$OUTDIR:/go/out" "$IMAGE"

echo ""
echo "Build complete! Binary at: $OUTDIR/mq_top"
echo ""
echo "To run (from an IBM MQ container):"
echo "  docker cp $OUTDIR/mq_top mqcontainer:/tmp/"
echo "  docker exec -it mqcontainer /tmp/mq_top -ibmmq.queueManager QM1"
