#!/bin/sh
# Dump Postgres format custom. Restore: pg_restore --clean --if-exists -d "$TARGET_URL" file.dump
set -eu
# .env lokal memakai DB_URL; DATABASE_URL tetap didukung (Railway).
DATABASE_URL="${DATABASE_URL:-${DB_URL:-}}"
: "${DATABASE_URL:?DATABASE_URL atau DB_URL belum di-set}"
out="backup-$(date +%Y%m%d).dump"
pg_dump --format=custom --no-owner --file="$out" "$DATABASE_URL"
echo "ok: $out"
