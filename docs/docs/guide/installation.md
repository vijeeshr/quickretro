---
title: Installing QuickRetro
description: Install QuickRetro in minutes using Docker, Docker Compose, or from source.
outline: deep
head:
  - - meta
    - name: keywords
      content: QuickRetro installation, install QuickRetro, QuickRetro Docker, QuickRetro Docker Compose, self-hosted retrospective tool, open source retrospective tool, agile retrospective software, retrospective board installation, QuickRetro setup, QuickRetro deploy
---

# Installation

The recommended way to install QuickRetro is by using Docker.

Ensure you have Docker Desktop OR Docker Engine with Compose plugin ready, and choose any of the below installation choices.

## Quick Install (using pre-built image)

This creates containers for the `app` (with the latest [`vijeesh82/quickretro-app:latest`](https://hub.docker.com/r/vijeesh82/quickretro-app/tags?name=latest) image from DockerHub), and `redis`, using the [`compose.install.yml`](https://github.com/vijeeshr/quickretro/blob/main/compose.install.yml) file.

### Linux / MacOS / Windows (with git bash)

```sh
# Download compose file to local directory
curl -LO https://raw.githubusercontent.com/vijeeshr/quickretro/main/compose.install.yml

# Run in detached mode in background
docker compose -f compose.install.yml up -d
```

### Windows PowerShell

```powershell
# Download compose file to local directory
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/vijeeshr/quickretro/main/compose.install.yml" -OutFile "compose.install.yml"

# Run in detached mode in background
docker compose -f compose.install.yml up -d
```

Visit http://localhost:8921 to open the app.

## Install (build from source)

This creates containers for the `app` (with image built from source code), and `redis`, using the [`compose.yml`](https://github.com/vijeeshr/quickretro/blob/main/compose.yml) file.

```sh
# Clone the repo in a empty directory
git clone https://github.com/vijeeshr/quickretro.git
cd quickretro

# Run in detached mode in background
docker compose up -d
```

Visit http://localhost:8921 to open the app.

## Install Full-Stack (build from source)

This creates containers for the `app` (with image built from source code), `redis`, and `caddy` as reverse-proxy using the [`compose.local.yml`](https://github.com/vijeeshr/quickretro/blob/main/compose.local.yml) file.

This is closest to production setup.

```sh
# Clone the repo in a empty directory
git clone https://github.com/vijeeshr/quickretro.git
cd quickretro

# Run in detached mode in background
docker compose -f compose.local.yml up -d
```

Visit https://localhost to open the app.

---

::: tip
For running QuickRetro directly in the host machine and outside of docker containers, check the [Development](development) page.
:::
