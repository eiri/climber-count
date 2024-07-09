# Climber Count

This is a Telegram bot that scrapes the rockgympro.com climber count gauge for a given gym and returns the latest people count to a channel when asked.

## How it works

- Periodically pulls HTML from rockgympro.com for a given `pgk` and `fId`.
- Parses the HTML file and extracts `var data` as JSON.
- Retrieves the counter for a given gym and stores it along with the update time in SQLite.
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
STORAGE - A path to the SQLite file.
BOT_TOKEN - A Telegram bot token from @BotFather.
```

For Docker, it is probably more convenient to use [docker-compose](compose.yaml).

## Licence

[MIT](https://github.com/eiri/climber-count/blob/main/LICENSE)
