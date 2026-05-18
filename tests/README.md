## Running e2e tests for backend

Navigate to root directory, and start local instance

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
go test -count=1 -v ./scenarios -run TestConnectHandshake
```
