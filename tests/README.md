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

Run tests (ignore cache)
```sh
cd ./tests/e2e/
go test -count=1 -v ./scenarios/...
```