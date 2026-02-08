## Running e2e tests for backend

If you are starting from scratch, build the frontend before running docker compose  
```sh
cd ./src/frontend/
npm install
npm run build-dev
```

Navigate back to root directory, and start local instance  
```sh
docker compose up
```

Run tests 
```sh
cd ./tests/e2e/
go test -v ./scenarios/...
```

Run tests (ignore cache)
```sh
cd ./tests/e2e/
go test -count=1 -v ./scenarios/...
```