---
title: Self-Hosting QuickRetro with Docker
description: Step-by-step guide to self-host QuickRetro using Docker. Learn configuration, deployment, and production setup.
outline: deep
head:
  - - meta
    - name: keywords
      content: self-host sprint retrospective, docker agile app, redis retro app
---

# Self Hosting

Although the [demo app](https://demo.quickretro.app) has all the features and can be used as-is, it runs on low resources. The data is auto-deleted within 2 days. It is recommended to self-host the app for better flexibility.

## Update Allowed-Origins

As defined in [Configurations](configurations#allowed-origins), update the config setting with your site origin.

## Secure Redis Instance

It is recommended to secure your Redis instance, preferably with ACL enabled. Check out the `redis` directory, and sample docker compose files `compose.yml`, `compose.local.yml`, `compose.prod.yml` etc in [github repository](https://github.com/vijeeshr/quickretro) for more details.

## Using Custom Config.toml

The default `config.toml` is baked right into the docker images. To use your config.toml with custom values, bind mount the toml file from host to your container.

`compose.yml`, `compose.local.yml` and `compose.prod.yml` in the repo already have the `volumes:` section, but is commented out by default.

Below steps are for `compose.yml`

```sh
# Stop and remove existing compose created items.
# Skip this if starting fresh.
docker compose down --rmi "all" --volumes
```

```yml
# For the "app" service in compose.yml file
# Uncomment "volumes:" section, or add if doesn't exist.
# Ensure custom config.toml is in the correct path in your host.
volumes:
  - ./src/config.toml:/app/config.toml:ro
```

```sh
# Run with volume mount
docker compose up

# To update config anytime later.
# Make the change in the bound config.toml,
# and restart app container to apply updated config
docker compose restart app
```

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

Check out the sample docker compose files `compose.yml`, `compose.local.yml`, `compose.prod.yml` etc in [github repository](https://github.com/vijeeshr/quickretro) for more details.

## Frequently Asked Questions

### Why isn't my `.env` file working on Linux?

Ensure it is saved with UTF-8 encoding and LF (Line Feed) line endings, not Windows CRLF. Creating it via standard Windows command prompt can cause encoding issues Docker Compose cannot read.
