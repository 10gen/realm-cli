# MongoDB Stitch CLI

## Installation

```
npm install mongodb-stitch-cli
npm install -g mongodb-stitch-cli
```

#### Building Binaries from Source
Build binary in place:
```
go build
```
Or, to build binary and install in your Go workspace's bin directory:
```
go install
```

To upload dependencies, the transpiler needs to be built as well:
```
cd etc/transpiler
yarn && yarn build
```

## Documentation

https://docs.mongodb.com/stitch/import-export/stitch-cli-reference/

#### When Using with a Local Stitch Server
When using `stitch-cli` against a locally running Stitch Server you can use any of the commands documented in the link above, however, you will need to pass some additional flags to the `stitch-cli` command in order for it to work properly.

##### `--base-url`
For all local commands use the `--base-url` flag to specify the URL of your locally running Stitch Server instance, e.g.:
```
stitch-cli import --base-url=http://localhost:8080
```

Where `http://localhost:8080` is the URL of your locally running instance. This flag is required for any command you want to run against your local server.

##### `--auth-provider=local-userpass`
When using `stitch-cli login` you will also need to include `--auth-provider=local-userpass` to authenticate with the local server using a username/password instead of the usual API Key method, e.g.:
```
stitch-cli login --base-url=http://localhost:8080 --auth-provider=local-userpass --username=USERNAME --password=PASSWORD
```

Where `USERNAME` and `PASSWORD` are the credentials for an existing local user.

## Linting

provided by gometalinter

```go
go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./...
```

## Testing

Run all tests:

```go
go test -v $(go list github.com/10gen/stitch-cli/...)
```


### Mocks

A custom mock of `StitchClient` can be found in `utils/test/utils.go` which can be used for simple mocking of most `StitchClient` methods.

If you need more sophisticated mocking utilities (such as being able to mock calls to the same method more than once in a single test) you can use the [`gomock`](https://github.com/golang/mock) version found in `api/mocks/stitch_api.go`

Run:

```
go run github.com/golang/mock/mockgen -source ./api/stitch_client.go -destination ./api/mocks/stitch_client.go
```
