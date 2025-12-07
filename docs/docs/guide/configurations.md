# Configurations
The application's default behaviour can be altered with configuration settings. This document provides a quick overview about it.

## Auto-Delete Duration
By default, data is deleted within 2 days in Redis. This can be updated by making the below changes.\
In the <code>src/config.toml</code> file, update the value for <code>auto_delete_duration</code>

```toml{5}
[data]
# Format: <number><unit>
# Units: s=seconds, m=minutes, h=hours, d=days
# Examples: "50s" for 50 seconds, "5m" for 5 minutes, "2h" for 2 hours, "7d" for 7 days
auto_delete_duration = "2d"
```

## Websocket Max Message Size
QuickRetro uses Websockets for communication. This configuration setting controls the max allowed size in bytes for all data sent through the websocket.

In the <code>src/config.toml</code> file, update the value for <code>max_message_size_bytes</code>
```toml{4}
[websocket]
# Maximum message size (in bytes) allowed from peer for the websocket connection
# For the front-end validation, keep the same value in (src/frontend/.env [VITE_MAX_WEBSOCKET_MESSAGE_SIZE_BYTES])
max_message_size_bytes = 1024
```

This setting is defined separately for the backend and frontend. For the frontend, this is defined in <code>src/frontend/.env</code>.\
Update the value for <code>VITE_MAX_WEBSOCKET_MESSAGE_SIZE_BYTES</code>
```ini{6}
VITE_WS_PROTOCOL=wss
VITE_SHOW_CONSOLE_LOGS=false
# Triggers message size validation.
# It is recommended to keep the same value as what's allowed in backend server (defined in src/config.toml [websocket].max_message_size_bytes).
# To avoid message size validation, comment out below line. However, this will break the server websocket connection when the limit is breached.
VITE_MAX_WEBSOCKET_MESSAGE_SIZE_BYTES=1024
```
::: danger IMPORTANT
Ensure the config values are same for both frontend and backend
:::

::: tip
<code>VITE_MAX_WEBSOCKET_MESSAGE_SIZE_BYTES</code> also causes UI validation to run everytime a User type's or paste's text.\
Commenting it out will stop the validation from being run everytime.\
It is not recommended to comment out this config, unless its causing issues for users.
:::

## Max Category Text Length
Available from <Badge type="tip" text="v1.6.0" />  

You can change the max number of characters allowed for each column name. Default is 80.

In the <code>src/config.toml</code> file, update the value for <code>max_category_text_length</code>
```toml{4}
[server]
# Maximum number of characters allowed for each category name
# For the front-end validation, keep the same value in (src/frontend/.env [VITE_MAX_CATEGORY_TEXT_LENGTH])
max_category_text_length = 80
```

This setting is defined separately for the backend and frontend. For the frontend, this is defined in <code>src/frontend/.env</code>.\
Update the value for <code>VITE_MAX_CATEGORY_TEXT_LENGTH</code>
```ini{3}
# Maximum number of characters allowed for each category name
# It is recommended to keep the same value as what's allowed in backend server (defined in src/config.toml [server].max_category_text_length).
VITE_MAX_CATEGORY_TEXT_LENGTH=80
```
::: danger IMPORTANT
Ensure the config values are same for both frontend and backend. 

Changing this also impacts the value defined in previous [Websocket Max Message Size](configurations#websocket-max-message-size) section.
Ensure that whatever value is set, **the websocket message/payload size doesn't exceed from what has been defined in previous section.**
:::

## Allowed Origins
Update the <code>allowed_origins</code> config setting in <code>src/config.toml</code> to add some degree of protection to the websocket connection.\
You will typically update this setting when [self-hosting](self-hosting).
```toml{7-14}
[server]
# When self-hosting, add your domain to allowed_origins list. 
# For e.g. if you are hosting your site at https://example.com, allowed_origins will look like -
# allowed_origins = [
#     "https://example.com"
# ]
allowed_origins = [
    "http://localhost:8080",
    "https://localhost:8080",
    "http://localhost:5173",
    "https://localhost",
    "https://quickretro.app",
    "https://demo.quickretro.app"
]
```

## Connecting to Redis
The Go app always attempts to connect to Redis when its starts. It errors out if connecting to Redis fails.
The app looks for an <code>ENV</code> variable named <code>REDIS_CONNSTR</code> for the connection details.

The Redis ACL username and password can be passed as part of the url to <code>REDIS_CONNSTR</code>. 

## Enable Cloudflare Turnstile
Turnstile is a smart CAPTCHA alternative from Cloudflare used to prevent bots. It is disabled by default for the Create board page.

To enable it, set the <code>TURNSTILE_ENABLED</code>, <code>TURNSTILE_SITE_KEY</code> and <code>TURNSTILE_SECRET_KEY</code> environment variables.

```ini{2-4}
REDIS_CONNSTR=<YOUR_REDIS_CONNECTION_STRING>
TURNSTILE_ENABLED=true
TURNSTILE_SITE_KEY=<YOUR_SITE_KEY>
TURNSTILE_SECRET_KEY=<YOUR_SECRET_KEY>
```

::: tip
You need to register with Cloudflare to get <code>TURNSTILE_SITE_KEY</code> and <code>TURNSTILE_SECRET_KEY</code>. Visit [Cloudflare](https://www.cloudflare.com/en-in/application-services/products/turnstile/) for more details.
:::
