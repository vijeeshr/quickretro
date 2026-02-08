# Configurations
The application's default behaviour can be altered with configuration settings. This document provides a quick overview about it.

## Websocket Max Message Size
QuickRetro uses Websockets for communication. This configuration setting controls the max allowed size in bytes for all data sent through the websocket.

In <code>src/config.toml</code>, update the value for <code>max_message_size_bytes</code>
```toml{3}
[websocket]
# Maximum message size (in bytes) allowed from peer for the websocket connection (also used by frontend)
max_message_size_bytes = 1024
```

## Auto-Delete Duration
By default, data is deleted within 2 days in Redis. This can be changed by making the below modification.\
In the <code>src/config.toml</code> file, update the value for <code>auto_delete_duration</code>

```toml{5}
[data]
# Format: <number><unit>
# Units: s=seconds, m=minutes, h=hours, d=days
# Examples: "50s" for 50 seconds, "5m" for 5 minutes, "2h" for 2 hours, "7d" for 7 days
auto_delete_duration = "2d"
```

## Max Category Text Length and Max Text Length
Available from <Badge type="tip" text="v1.6.0" /><Badge type="tip" text="v1.6.3" /> 

In <code>src/config.toml</code> file, control the number of characters allowed for each column name by setting <code>max_category_text_length</code>   
Set the limit for number of characters allowed for Board name, Team name, Nickname with <code>max_text_length</code>  
Default is 80 for both.
```toml{3,5}
[data]
# Maximum number of characters allowed for each category name (also used by frontend)
max_category_text_length = 80
# Maximum number of characters allowed for board name, team name, nickname (also used by frontend)
max_text_length = 80
```
::: danger IMPORTANT
Changing this also impacts the value defined in [Websocket Max Message Size](configurations#websocket-max-message-size) section.
Ensure that whatever value is set, **the websocket message/payload size doesn't exceed from what has been defined in mentioned section.**
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

## Running in a different Port
By default, the app starts at port <code>8080</code>  
To run in a different port, for e.g. 9090, change the value of <code>ENV</code> variable named <code>PORT</code> in <code>.env</code> file
```ini{1}
PORT=9090
REDIS_CONNSTR=<YOUR_REDIS_CONNECTION_STRING>
TURNSTILE_ENABLED=true
TURNSTILE_SITE_KEY=<YOUR_SITE_KEY>
TURNSTILE_SECRET_KEY=<YOUR_SECRET_KEY>
```
Whitelist the new port, by changing the default port (_or adding new entries in the list_) in the [Allowed Origins](configurations#allowed-origins) section in the <code>src/config.toml</code> file
```toml{8-9}
[server]
# When self-hosting, add your domain to allowed_origins list. 
# For e.g. if you are hosting your site at https://example.com, allowed_origins will look like -
# allowed_origins = [
#     "https://example.com"
# ]
allowed_origins = [
    "http://localhost:9090",
    "https://localhost:9090",
    "http://localhost:5173",
    "https://localhost",
    "https://quickretro.app",
    "https://demo.quickretro.app"
]
```
This is enough to run the app in port 9090 using docker compose.  

::: tip Direct Frontend to connect to new port during development
During development, you may need to run the Vue frontend outside the docker container. While the above change works when the app is built and run entirely in a docker container with Docker Compose, the Vue frontend won't automatically connect to the newly changed port 9090 running inside the docker container.  

Change value of VITE env variable <code>VITE_API_BASE_URL</code> in <code>src/frontend/.env</code> to make the frontend(_running outside of docker_) connect to backend(_running inside docker container with new port 9090 exposed to host_).  
```ini{3}
VITE_SHOW_CONSOLE_LOGS=false
VITE_TURNSTILE_SCRIPT_URL=https://challenges.cloudflare.com/turnstile/v0/api.js
VITE_API_BASE_URL=http://localhost:9090
```
:::

::: tip
The repo has some examples with CaddyFile, which uses 8080. If using CaddyServer, remember to update it to 9090 as well.
:::

## Connecting to Redis
The Go app always attempts to connect to Redis when its starts. It errors out if connecting to Redis fails.
The app looks for an <code>ENV</code> variable named <code>REDIS_CONNSTR</code> for the connection details.

The Redis ACL username and password can be passed as part of the url to <code>REDIS_CONNSTR</code>. 

## Enable Cloudflare Turnstile
Turnstile is a smart CAPTCHA alternative from Cloudflare used to prevent bots. It is disabled by default for the Create board page.

To enable it, set the <code>TURNSTILE_ENABLED</code>, <code>TURNSTILE_SITE_KEY</code> and <code>TURNSTILE_SECRET_KEY</code> environment variables.

```ini{3-5}
PORT=8080
REDIS_CONNSTR=<YOUR_REDIS_CONNECTION_STRING>
TURNSTILE_ENABLED=true
TURNSTILE_SITE_KEY=<YOUR_SITE_KEY>
TURNSTILE_SECRET_KEY=<YOUR_SECRET_KEY>
```

::: tip
You need to register with Cloudflare to get <code>TURNSTILE_SITE_KEY</code> and <code>TURNSTILE_SECRET_KEY</code>. Visit [Cloudflare](https://www.cloudflare.com/en-in/application-services/products/turnstile/) for more details.
:::

## Content limit notification delay
Available from <Badge type="tip" text="v1.6.3" /> 

Set a reasonable value to <code>content_editable_invalid_debounce_ms</code> in <code>src/config.toml</code> to prevent the user from getting bombarded with notification messages on every keystroke, when content limit is breached in Cards or Comments

```toml{3}
[frontend]
# Delay (in milliseconds) before showing "message size limit reached" notification for cards/comments
content_editable_invalid_debounce_ms = 500
```

## Typing indicators
Available from <Badge type="tip" text="v1.6.3" />  

Enable and control a "_User is typing_" indicator that appears as a pulse animation on a User's avatar in the Dashboard's right panel with the below settings.  

What each setting does is defined in comments in <code>src/config.toml</code>

```toml{7,11,16,19}
# ---------------------------------------------------------------------------------------------------
# Typing Activity Notifications
# Controls the real-time "user is typing" activity indicator shown on avatars during a retro session.
# ---------------------------------------------------------------------------------------------------
[typing_activity]
# Enable or disable typing activity notifications globally (also used by frontend)
enabled = true
# Automatically disable dispatch of "typing" events in the frontend when 
# the number of other active users(excluding current user) in a board exceeds this value.
# Set to 0 (or a negative value) to disable this limit. (only used by frontend)
auto_disable_after_count = 15
# Minimum time (in milliseconds) between consecutive "typing" events emitted by the same client. (only used by frontend)
# This acts as a throttle to:
# - Reduce WebSocket noise
# - Prevent excessive broadcasts while a user is actively typing for a long time
emit_throttle_ms = 3000
# Time (in milliseconds) after which the typing indicator is automatically cleared if no new typing event is received. (only used by frontend)
# This controls how long the activity ring stays visible on the user's avatar after they stop typing.
display_timeout_ms = 2000
```
