![banner](banner.png)

moni.gg is a status monitor written in golang.

## Usage
The status monitor gets executed every x minutes via cron. All monitoring targets have their own file inside the `config/` folder (gets created upon initial execution). It currently supports only http(s) and mysql/mariadb monitoring.

[Example Statuspage](https://status.opfw.net/)

### http(s)
The config files for http(s) targets are standard `.http` files that you can export from postman.

![postman](https://i.shrt.day/LOCuCiKo93.png)

### mysql/mariadb
The config files for mysql/mariadb targets have the `.mysql` extension. The first line of the file contains connection information like so:
```
Hostname=localhost;Username=root;Password=password1234;Database=my-db;Port=3306
```
The second line is optional and contains the query that should be executed to test database connectivity. This query should always return 1 or more rows. If the query returns 0 rows, the target is considered down. It defaults to a simple `SELECT 1`.

## Configuration
Configuration of the monitor is done via the `.env` file.
```env
# URL of your status page (used for email notifications)
STATUS_PAGE=https://status.example.com

# Optional: Email target (when a target goes down/up)
EMAIL_TO=hello@example.com

# Optional: SMTP configuration for sending emails
SMTP_HOST=
SMTP_USER=
SMTP_PASSWORD=

# Path to your favicon
TEMPLATE_FAVICON=

# Path to your banner
TEMPLATE_BANNER=

# URL of the status page
TEMPLATE_URL=

# Title of the status page
TEMPLATE_TITLE=

# Description of the status page
TEMPLATE_DESCRIPTION=
```

After you've configured the `.env` file, you have to copy the `public/` directory to your webserver. The `public/` directory contains the status page that is shown to your users. It has to be in the same directory as the `config/` directory and the executable.

Once you've finished setting up the monitor, you can add it to your crontab. This example executes the monitor every 5 minutes.
```cron
*/5 * * * * /path/to/monitor
```

## API
The monitor offers a JSON api that can be used to retrieve an overview of all targets and their current status. The api is available at `/summary.json`. If you call the api from another domain, you'll have to set the `Access-Control-Allow-Origin` header. I'd also recommend disabling browser caching to avoid stale data.

**Nginx example:**
```nginx
location /summary.json {
    # Disable caching
    expires -1;

    # Allow cross-origin requests
    add_header Access-Control-Allow-Origin *;
}
```