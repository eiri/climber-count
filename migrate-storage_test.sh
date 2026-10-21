#!/usr/bin/env bash
set -euo pipefail

# Run only against an explicit test database; every case gets a private schema.
: "${TEST_POSTGRES_URL:?set TEST_POSTGRES_URL to a dedicated PostgreSQL test database}"
script=$(cd "$(dirname "$0")" && pwd)/migrate-storage.sh
tmp=$(mktemp -d)
schema="migration_test_$$"
source="$tmp/source ?#'.db"
export PGOPTIONS="-c search_path=$schema"

pg() {
  psql --dbname="$TEST_POSTGRES_URL" -X -w -q -At -v ON_ERROR_STOP=1 "$@"
}

cleanup() {
  pg -c "DROP SCHEMA IF EXISTS $schema CASCADE" >/dev/null || echo "test schema cleanup failed" >&2
  rm -rf "$tmp"
}
trap cleanup EXIT

reset_target() {
  pg -c "DROP SCHEMA IF EXISTS $schema CASCADE; CREATE SCHEMA $schema" >/dev/null
  if [ "$(pg -c 'SELECT current_schema()')" != "$schema" ]; then
    echo "test URL overrides the private search_path; refusing tests" >&2
    exit 1
  fi
}

new_source() {
  rm -f "$source"
  sqlite3 "$source" <<'SQL'
CREATE TABLE counter (id INTEGER PRIMARY KEY AUTOINCREMENT, gym TEXT NOT NULL, count INTEGER, capacity INTEGER, last_update TEXT);
CREATE TABLE visit (id INTEGER PRIMARY KEY AUTOINCREMENT, gym TEXT NOT NULL, timestamp TEXT, action TEXT);
SQL
}

run_import() {
  MIGRATE_POSTGRES_URL="$TEST_POSTGRES_URL" bash "$script" --postgres "$source"
}

assert_equal() {
  if [ "$1" != "$2" ]; then
    printf 'got <%s>, expected <%s>\n' "$1" "$2" >&2
    exit 1
  fi
}

expect_failure() {
  if "$@" >"$tmp/error" 2>&1; then
    echo "expected failure" >&2
    exit 1
  fi
}

reset_target
new_source
sqlite3 "$source" <<'SQL'
INSERT INTO counter VALUES (1, 'O''THER', NULL, NULL, ''), (2, 'TST', 0, 100, '2024-05-30T10:00:00+02:00'), (10, 'TST', 3, 100, '2024-05-30T10:00:00Z');
INSERT INTO visit VALUES (7, 'TST', '2024-05-30T10:00:00Z', 'in'), (12, 'TST', '2024-05-30T10:00:00Z', 'out'),
(9, 'SBG', NULL, 'comma,"quote"' || char(10) || '\.' || char(10) || 'é' || char(13));
SQL
before=$(cksum "$source")
run_import
assert_equal "$(cksum "$source")" "$before"
assert_equal "$(pg -c 'SELECT COUNT(*) FROM counter')" 3
assert_equal "$(pg -c 'SELECT COUNT(*) FROM visit')" 3
assert_equal "$(pg -c "SELECT (count IS NULL AND capacity IS NULL AND last_update = '') FROM counter WHERE id=1")" t
assert_equal "$(pg -c "SELECT count FROM counter WHERE gym='TST' ORDER BY id DESC LIMIT 1")" 3
assert_equal "$(pg -c "SELECT action FROM visit WHERE gym='TST' ORDER BY timestamp DESC, id DESC LIMIT 1")" out
assert_equal "$(pg -c "SELECT encode(convert_to(action, 'UTF8'), 'hex') FROM visit WHERE id=9")" \
  "$(sqlite3 "$source" 'SELECT lower(hex(action)) FROM visit WHERE id=9')"
assert_equal "$(pg -c "INSERT INTO counter (gym) VALUES ('NEW') RETURNING id")" 11
assert_equal "$(pg -c "INSERT INTO visit (gym) VALUES ('NEW') RETURNING id")" 13
expect_failure run_import
grep -q 'not empty' "$tmp/error"
assert_equal "$(pg -c 'SELECT COUNT(*) FROM counter')" 4

# Verification must detect silent changes, then roll back data and sequences.
reset_target
run_import
pg -c 'TRUNCATE counter, visit; ALTER TABLE counter ALTER COLUMN id RESTART WITH 80' >/dev/null
pg <<'SQL'
CREATE FUNCTION change_counter() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN NEW.count := COALESCE(NEW.count, 0) + 1; RETURN NEW; END $$;
CREATE TRIGGER change_counter BEFORE INSERT ON counter FOR EACH ROW EXECUTE FUNCTION change_counter();
SQL
expect_failure run_import
grep -q 'verification failed' "$tmp/error"
assert_equal "$(pg -c 'SELECT COUNT(*) FROM counter')" 0
assert_equal "$(pg -c 'SELECT COUNT(*) FROM visit')" 0
pg -c 'DROP TRIGGER change_counter ON counter' >/dev/null
assert_equal "$(pg -c "INSERT INTO counter (gym) VALUES ('NEW') RETURNING id")" 80

# A failure after copying counters must undo that copy and its sequence restart.
reset_target
new_source
sqlite3 "$source" "INSERT INTO counter (id, gym) VALUES (10, 'TST'); INSERT INTO visit (id, gym) VALUES (9223372036854775807, 'TST');"
expect_failure run_import
assert_equal "$(pg -c "SELECT COUNT(*) FROM pg_class WHERE relnamespace='$schema'::regnamespace AND relkind IN ('r', 'S')")" 0

new_source
sqlite3 "$source" "INSERT INTO counter (id, gym) VALUES (10, 'TST');"
run_import
pg -c 'TRUNCATE counter, visit; ALTER TABLE counter ALTER COLUMN id RESTART WITH 80' >/dev/null
sqlite3 "$source" "INSERT INTO visit (id, gym) VALUES (9223372036854775807, 'TST');"
expect_failure run_import
assert_equal "$(pg -c 'SELECT COUNT(*) FROM counter')" 0
assert_equal "$(pg -c "INSERT INTO counter (gym) VALUES ('NEW') RETURNING id")" 80
reset_target

# Missing tables, empty sources, blobs, NUL bytes, and integer overflow must fail.
new_source
sqlite3 "$source" 'DROP TABLE visit'
expect_failure run_import
new_source
expect_failure run_import
new_source
sqlite3 "$source" "INSERT INTO counter (gym) VALUES (char(0));"
expect_failure run_import
new_source
sqlite3 "$source" "INSERT INTO counter (gym, last_update) VALUES ('TST', X'4142');"
expect_failure run_import
new_source
sqlite3 "$source" "INSERT INTO counter (gym, count) VALUES ('TST', 2147483648);"
expect_failure run_import
assert_equal "$(pg -c "SELECT COUNT(*) FROM pg_class WHERE relnamespace='$schema'::regnamespace AND relkind IN ('r', 'S')")" 0

# One empty table is valid; the other table still blocks repeat imports.
new_source
sqlite3 "$source" "INSERT INTO visit (id, gym) VALUES (20, 'TST');"
run_import
assert_equal "$(pg -c "INSERT INTO counter (gym) VALUES ('NEW') RETURNING id")" 1
assert_equal "$(pg -c "INSERT INTO visit (gym) VALUES ('NEW') RETURNING id")" 21
pg -c 'TRUNCATE counter' >/dev/null
expect_failure run_import
grep -q 'not empty' "$tmp/error"

expect_failure env MIGRATE_POSTGRES_URL="$TEST_POSTGRES_URL" bash "$script" --postgres "$tmp/missing.db"
assert_equal "$(test -e "$tmp/missing.db" && echo exists || echo missing)" missing
expect_failure env MIGRATE_POSTGRES_URL="$TEST_POSTGRES_URL" bash "$script" --postgres "$tmp"
expect_failure env MIGRATE_POSTGRES_URL=sqlite://wrong bash "$script" --postgres "$source"
expect_failure bash "$script" --postgres

# Keep the original per-gym SQLite merge working.
legacy="$tmp/sbl.db"
sqlite3 "$legacy" <<'SQL'
CREATE TABLE count (id INTEGER PRIMARY KEY, count INTEGER, capacity INTEGER, last_update TEXT);
CREATE TABLE gym (id INTEGER PRIMARY KEY, timestamp TEXT, action TEXT);
INSERT INTO count VALUES (1, 7, 100, '2024-05-30T10:00:00Z');
INSERT INTO gym VALUES (1, '2024-05-30T10:00:00Z', 'in');
SQL
bash "$script" "$legacy" "$tmp/merged.db"
assert_equal "$(sqlite3 "$tmp/merged.db" 'SELECT gym || count FROM counter')" SBL7
assert_equal "$(sqlite3 "$tmp/merged.db" 'SELECT gym || action FROM visit')" SBLin
printf 'All storage migration tests passed.\n'
