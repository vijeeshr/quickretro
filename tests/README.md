## Running e2e tests for backend

Start local instance with docker compose from root directory  
```sh
docker compose up
```

Run tests 
```sh
cd ./tests/e2e/
go test -v ./scenarios/...
```