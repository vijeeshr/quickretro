# quickretro
A websocket based example app for conducting a quick sprint retro.

## Live app demo
[https://quickretro.app](https://quickretro.app){:target="_blank"}

## Runnning the app locally
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
npm run build-dev
```

### To start the Golang backend server
Navigate back to root directory.
```sh
docker compose up
```
Visit http://localhost:8080 to open the Vue app and start creating a board.  

## For Frontend Development
### Runing Vue app in development mode
Run the app.  
```sh
npm run dev
```
Visit http://localhost:5173/ to open.  

## Features/Capabilities/Limitations
No logins.  
Share the board url with people to participate in the retro meeting.  
Board creator can opt to blur/unblur messages during the retro meeting.  
Board creator can update/delete any message (including messages of other participants).   
Board creator can download details as Pdf.  
Online presence display.   
All data auto-deleted within 2 hours of most recent update.  

![quickretro1](https://github.com/vijeeshr/quickretro/assets/16733867/020b40d8-5b11-4daf-a2f3-95a0ee17f918)

![quickretro2](https://github.com/vijeeshr/quickretro/assets/16733867/6802b697-362b-4f99-b6da-8b9bc0c3c4ab)


## Note
Use in production after vetting it. This example was created with the intent of learning Go.
