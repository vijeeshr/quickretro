# quickretro
A websocket based example app for conducting a quick sprint retro.

## Runnning the app
Ensure Go, Nodejs and Docker are installed.  

### Build the Vue frontend
```sh
cd .\src\frontend\
```
Install packages and dependencies.  
```sh
npm install
```
Build the frontend.  
This creates assets in "frontend/dist" directory. This dist directory is embedded in the backend Golang binary.  
```sh
npm run build
```

### To start the Golang backend server
Navigate back to root directory.
```sh
docker compose up
```
Visit http://localhost:8080 to open the Vue app and start creating a board.  

## For Development
### Runing Vue app in development mode
Run the app.  
```sh
npm run dev
```
Visit http://localhost:5173/ to open.  

## Features
No logins.  
Share the board url with people to participate in the retro meeting.  
Online presence display.  
Board creator can opt to blur/unblur messages during the retro meeting.  

![quickretro1](https://github.com/vijeeshr/quickretro/assets/16733867/020b40d8-5b11-4daf-a2f3-95a0ee17f918)

![quickretro2](https://github.com/vijeeshr/quickretro/assets/16733867/6802b697-362b-4f99-b6da-8b9bc0c3c4ab)


## Note
Do not use this in production. This example was created with the intent of learning Go.
