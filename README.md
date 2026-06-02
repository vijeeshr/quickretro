# quickretro

A websocket based app for conducting a quick sprint retro.

<img width="1797" height="942" alt="QuickRetro Dashboard" src="https://github.com/user-attachments/assets/12145787-6971-41a3-9455-5415fbbb3567" />

## Live app demo

https://demo.quickretro.app

## Site

https://quickretro.app

### Docs

https://quickretro.app/guide/getting-started

## Installation

The recommended way to install QuickRetro is by using Docker.

Ensure you have Docker Desktop OR Docker Engine with Compose plugin ready, and choose any of the below installation choices.

### Quick Install (using pre-built image)

This creates containers for the `app` (with the latest [`vijeesh82/quickretro-app:latest`](https://hub.docker.com/r/vijeesh82/quickretro-app/tags?name=latest) image from DockerHub), and `redis`, using the [`compose.install.yml`](https://github.com/vijeeshr/quickretro/blob/main/compose.install.yml) file.

#### Linux / MacOS / Windows (with git bash)

```sh
# Download compose file to local directory
curl -LO https://raw.githubusercontent.com/vijeeshr/quickretro/main/compose.install.yml

# Run in detached mode in background
docker compose -f compose.install.yml up -d
```

#### Windows PowerShell

```powershell
# Download compose file to local directory
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/vijeeshr/quickretro/main/compose.install.yml" -OutFile "compose.install.yml"

# Run in detached mode in background
docker compose -f compose.install.yml up -d
```

Visit http://localhost:8921 to open the app.

To install by building from source, refer the [installation docs](https://quickretro.app/guide/installation)

## Setup for local development

Ensure Go, Nodejs and Docker are installed.

### Build the Vue frontend

```sh
cd ./src/frontend/
```

Install packages and dependencies.

```sh
npm install
```

Build the frontend.  
This creates assets in "frontend/dist" directory. This dist directory is embedded in the backend Golang binary.

```sh
npm run build-dev
```

### To start the Golang backend server

Navigate back to root directory.

```sh
docker compose up
```

Visit http://localhost:8921 (for older versions, use http://localhost:8080) to open the Vue app and start creating a board.

## For Frontend Development

### Running Vue app in development mode

Run the app.

```sh
npm run dev
```

Visit http://localhost:5173/ to open.

## Features

- No Signups or Logins - That's right! No need to signup or login.
- No Board Limits - Create Boards or Invite Users without limits.
- Mobile Friendly UI - Easily participate from your mobile phone.
- Customize Column Names - Choose upto 5 columns with any name.
- Mask/Blur messages - Avoid revealing messages of other participants.
- Anonymous Messages - Post messages without revealing your name.
- Download as PDF - Download messages as PDF.
- Countdown Timer - Stopwatch with max 1 hour limit.
- Board Lock - Lock to stop addition/updation of messages.
- Dark Theme - Easily switch to use a Dark theme.
- Online Presence Display - See participants present in the meeting in realtime.
- Auto-Delete data - Auto-delete with configurable retention duration.
