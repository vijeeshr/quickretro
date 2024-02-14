# quickretro
A websocket based example app for conducting a quick sprint retro.

## Runnning the app
### To start the Golang backend server
```sh
docker compose up
```
### To start the Vue frontend
```sh
cd .\frontend\quickretroapp\
```
```sh
npm run dev
```
Visit http://localhost:5173/ to open the Vue app and start creating a board.  
Can also visit http://localhost:8080, to view the Html JS only UI. That won't be updated going forward and may be removed.  

## Features
No logins.  
Share the board url with people to participate in the retro meeting.  
Online presence display.  
Board creator can opt to blur/unblur messages during the retro meeting.  

![quickretro1](https://github.com/vijeeshr/quickretro/assets/16733867/020b40d8-5b11-4daf-a2f3-95a0ee17f918)

![quickretro2](https://github.com/vijeeshr/quickretro/assets/16733867/6802b697-362b-4f99-b6da-8b9bc0c3c4ab)


## Note
Do not use this in production. This example was created with the intent of learning Go.
