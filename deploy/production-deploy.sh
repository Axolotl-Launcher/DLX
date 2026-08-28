#!/usr/bin/env bash
set -euo pipefail

APP_DIR=/opt/axolotl-dlx
TAG="${1:?usage: production-deploy.sh <immutable-image-tag>}"
cd "$APP_DIR"

# Secrets are server-only files under ./secrets and are never fetched from GitHub.
export IMAGE_TAG="$TAG"
docker-compose -f compose.production.yaml -f compose.server.yaml -f compose.ci.yaml pull dlx sponsor-gateway sponsor-web
docker-compose -f compose.production.yaml -f compose.server.yaml -f compose.ci.yaml up -d --no-build dlx sponsor-gateway sponsor-web

for _ in $(seq 1 24); do
  if curl --fail --silent --show-error http://127.0.0.1:8080/readyz >/dev/null && curl --fail --silent --show-error http://127.0.0.1:8081/ >/dev/null; then
    docker-compose -f compose.production.yaml -f compose.server.yaml -f compose.ci.yaml ps
    exit 0
  fi
  sleep 5
done

echo "Gateway did not become ready" >&2
docker-compose -f compose.production.yaml -f compose.server.yaml -f compose.ci.yaml logs --tail=100 sponsor-gateway >&2
exit 1
