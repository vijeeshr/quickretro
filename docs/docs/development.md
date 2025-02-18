---
outline: deep
---
# Development Guide
This guide is intended to help you get started with running the application locally.
::: tip 
For development, ensure Go (ver >= 1.22.0), NodeJs and Docker are installed locally
:::

## Tech Stack
Go <Badge type="tip" text="1.22.0" /> for backend\
Vue <Badge type="tip" text="^3.3.11" /> for frontend\
Redis is used as the datastore and for pubsub
::: info
The Go app runs as a single binary with the frontend embedded inside it
:::

## Running locally
The easiest way to run locally is by using Docker.

### Build Vue frontend
The Vue frontend must be built first
```sh
cd .\src\frontend\
npm install
npm run build-dev
```
This installs the packages, dependencies, and creates assets in <code>frontend/dist</code> directory. This dist directory is embedded in the backend Golang binary when it is built.\

### Run with Docker
Navigate back to root directory and run -
```sh
docker compose up
```
This builds and starts a docker container for the app, and another container with Redis.\
The app starts at http://localhost:8080

## Setting up for Development
Ensure you have Redis running.
::: tip
From the previous docker step, you can keep the Redis container running and stop the other containers
:::
### Running Go backend app
To run the Go app directly (outside the container) -
Open a terminal and from the root directory
```sh
cd .\src\
go run .
```
This starts the Go server. You are ready to make changes to the Go app now.

::: tip
Go must be installed for this step.
:::

### Running Vue frontend app
Open another terminal and from the root directoy - 
```sh
cd .\src\frontend\
npm run dev
```
This starts the Vue app at http://localhost:5173\
Feel free to make changes to the app.