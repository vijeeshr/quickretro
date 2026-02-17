# Self-Hosting

Although the [demo app](https://demo.quickretro.app) has all the features and can be used as-is, it runs on low resources. The data is auto-deleted within 2 days. It is recommended to self-host the app for better flexibility.

## Update Allowed-Origins

As defined in [Configurations](configurations#allowed-origins), update the config setting with your site origin.

## Secure Redis Instance

It is recommended to secure your Redis instance, preferably with ACL enabled. Check out the `redis` directory, and sample docker compose files `compose.yml`, `compose.reverseproxy.yml`, `compose.demohosting.yml` etc in [github repository](https://github.com/vijeeshr/quickretro) for more details.

## Passing ENV variables with Compose

Environment variables are passed using `.env` file which is present in the same directory as `compose\*.yml` files.\
Example: Create an env file with your values -

```sh
echo "REDIS_CONNSTR=redis://redis:6379/0" > .env
# echo "MY_VAR1=false" >> .env
# echo "MY_VAR2=true" >> .env
```

::: info
To securely pass `ENV` vars, feel free to use an approach which suits you best.
:::
::: warning NOTE
DO NOT create the file directly from Windows `CMD` if you intend to run the app in Linux. It creates Unicode text, UTF-16, little-endian text, with CRLF line terminators. This causes problems for Docker Compose to read the env file.

On Windows, you can create the file in UTF-8 using Git Terminal.
:::

## Sample Compose files

Check out the sample docker compose files `compose.yml`, `compose.reverseproxy.yml`, `compose.demohosting.yml` etc in [github repository](https://github.com/vijeeshr/quickretro) for more details.
