# Climber Count

Scrapes the rockgympro.com climber count gauge.

## How it works

- Pulls html from rockgympro.com for a given uid and fId.
- Parses the html file, extracts `var data` as JSON.
- Gets a counter for a given gym and stores it and update time in csv file.
- If the count is less than a given theashold sends a message through a Telegram bot.
- Implements a webhook for the Telegram bot to answer a question "How many on the wall"?

To use set the cli in crontab for desired timeframes. Play nicely and don't spam URL with requests, once in quater of an hour sure should be enough.

### Config

Pass URL's uid and fID plus name of the gym either as flags or environment variables.

```bash
$ ./climber-count --help
Usage of ./climber-count:
  -fid string
    	URL's fId. (Defaults to env var CC_FID)
  -gym string
    	Gym  name. (Defaults to env var CC_GYM)
  -uid string
    	URL's uid. (Defaults to env var CC_UID)
```

For dev just create `.env` file with those vars and Makefile pass them to run.

## Licence

[MIT](https://github.com/eiri/climber-count/blob/main/LICENSE)
