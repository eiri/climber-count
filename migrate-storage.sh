#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $0 source.db... target.db" >&2
}

abs_path() {
  local path="$1"
  local dir base
  dir=$(dirname "$path")
  base=$(basename "$path")
  mkdir -p "$dir"
  dir=$(cd "$dir" && pwd -P)
  printf '%s/%s\n' "$dir" "$base"
}

sql_quote() {
  printf "%s" "$1" | sed "s/'/''/g"
}

run_sqlite() {
  sqlite3 -batch -init /dev/null "$@"
}

if ! command -v sqlite3 >/dev/null 2>&1; then
  echo "sqlite3 is required" >&2
  exit 1
fi

if [ "$#" -lt 2 ]; then
  usage
  exit 1
fi

target="${!#}"
sources=("${@:1:$#-1}")

target_abs=$(abs_path "$target")

for src in "${sources[@]}"; do
  if [ ! -r "$src" ]; then
    echo "source is not readable: $src" >&2
    exit 1
  fi

  src_abs=$(abs_path "$src")
  if [ "$src_abs" = "$target_abs" ]; then
    echo "target must not be a source: $target" >&2
    exit 1
  fi

  run_sqlite "$src" "PRAGMA schema_version;" >/dev/null
done

mkdir -p "$(dirname "$target")"
rm -f "$target"

run_sqlite "$target" <<'SQL'
CREATE TABLE counter (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  gym TEXT NOT NULL,
  count INTEGER,
  capacity INTEGER,
  last_update TEXT
);

CREATE TABLE visit (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  gym TEXT NOT NULL,
  timestamp TEXT,
  action TEXT
);

CREATE INDEX idx_counter_gym_id ON counter(gym, id);
CREATE INDEX idx_visit_gym_timestamp ON visit(gym, timestamp, id);
SQL

for src in "${sources[@]}"; do
  name=$(basename "$src")
  gym=${name%.*}
  gym=$(printf "%s" "$gym" | tr '[:lower:]' '[:upper:]')

  has_count=$(run_sqlite "$src" "SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'count';")
  has_gym=$(run_sqlite "$src" "SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'gym';")

  src_sql=$(sql_quote "$src")
  gym_sql=$(sql_quote "$gym")

  if [ "$has_count" = "1" ]; then
    run_sqlite "$target" <<SQL
ATTACH '$src_sql' AS src;
INSERT INTO counter (gym, count, capacity, last_update)
SELECT '$gym_sql', "count", capacity, last_update FROM src."count" ORDER BY id;
DETACH src;
SQL
  fi

  if [ "$has_gym" = "1" ]; then
    run_sqlite "$target" <<SQL
ATTACH '$src_sql' AS src;
INSERT INTO visit (gym, timestamp, action)
SELECT '$gym_sql', timestamp, action FROM src.gym ORDER BY id;
DETACH src;
SQL
  fi
done
