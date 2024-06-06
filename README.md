# Climber Count

This is a Telegram bot that scrapes the rockgympro.com climber count gauge for a given gym and returns the lates people count to a channel when asked.

## How it works

- Periodically pulls html from rockgympro.com for a given pgk and fId.
- Parses the html file, extracts `var data` as JSON.
- Gets a counter for a given gym and stores it and update time in csv file.
- When bot asked for /count returns the latest count from the storage.

## Licence

[MIT](https://github.com/eiri/climber-count/blob/main/LICENSE)
