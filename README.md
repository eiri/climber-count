# Climber Count

This is a Telegram bot that scrapes the rockgympro.com climber count gauge for a given gym and returns the latest people count to a channel when asked.

## How it works

- Periodically pulls HTML from rockgympro.com for a given `pgk` and `fId`.
- Parses the HTML file and extracts `var data` as JSON.
- Retrieves the counter for a given gym and stores it along with the update time in SQLite or PostgreSQL.
- When the bot is asked for `/count`, it returns the latest count from the storage.

## Installation

You can either download the binary from the release tab, use Homebrew with `brew tap eiri/tap` and `brew install climber-count`, or use Docker with `docker pull ghcr.io/eiri/climber-count`.

## Configuration

Export the following environment variables:

```
PGK - A rockgympro.com UID. This is usually an MD5-like string in the URL path set right after "public".
FID - Another ID, organization-specific.
GYM - A gym abbreviation. The response can contain multiple gyms' counters.
SCHEDULE - Key=crontab pairs separated by |. For example: weekdays=4 */5 8-22 * * MON-FRI|weekends=2 */5 8-20 * * SAT,SUN. This pulls the counter every five minutes during the gym's working hours. Theoretically, it can go down to seconds, but there is no need to spam rockgympro.com. Be nice.
STORAGE - SQLite or PostgreSQL URL. Defaults to sqlite://data/storage.db.
BOT_TOKEN - A Telegram bot token from @BotFather.
METRICS_ADDR - Optional Prometheus metrics listen address. Defaults to :9117. Metrics are exposed at /metrics.
```

For Docker, it is probably more convenient to use [docker-compose](docker-compose.yaml).

## Storage

SQLite is the default. PostgreSQL accepts `postgres://` and `postgresql://` URLs, for example `postgres://user:password@localhost:5432/climber_count?sslmode=require`.
The database must exist. The application creates its tables and indexes at startup.
Changing `STORAGE` does not copy existing data. Keep database credentials out of version control.

Tests use temporary SQLite files by default. To run the same tests against PostgreSQL, set `TEST_POSTGRES_URL`:

```sh
TEST_POSTGRES_URL='postgres://user:password@localhost:5432/climber_count_test?sslmode=disable' go test ./...
```

Use a separate test database. Its user must have permission to create schemas.
Each test creates a separate schema and removes it after the test.
Use `sslmode=disable` only for local test connections.

## Licence

[MIT](https://github.com/eiri/climber-count/blob/main/LICENSE)
