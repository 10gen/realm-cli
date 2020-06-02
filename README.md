# MongoDB Realm CLI

## Installation

```
npm install mongodb-realm-cli
npm install -g mongodb-realm-cli
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

https://docs.mongodb.com/realm/import-export/realm-cli-reference/

#### When Using with a Local Realm Server
When using `realm-cli` against a locally running Realm Server you can use any of the commands documented in the link above, however, you will need to pass some additional flags to the `realm-cli` command in order for it to work properly.

##### `--base-url`
For all local commands use the `--base-url` flag to specify the URL of your locally running Realm Server instance, e.g.:
```
realm-cli import --base-url=http://localhost:8080
```

Where `http://localhost:8080` is the URL of your locally running instance. This flag is required for any command you want to run against your local server.

##### `--auth-provider=local-userpass`
When using `realm-cli login` you will also need to include `--auth-provider=local-userpass` to authenticate with the local server using a username/password instead of the usual API Key method, e.g.:
```
realm-cli login --base-url=http://localhost:8080 --auth-provider=local-userpass --username=USERNAME --password=PASSWORD
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
go test -v $(go list github.com/10gen/realm-cli/...)
```


### Mocks

A custom mock of `RealmClient` can be found in `utils/test/utils.go` which can be used for simple mocking of most `RealmClient` methods.

If you need more sophisticated mocking utilities (such as being able to mock calls to the same method more than once in a single test) you can use the [`gomock`](https://github.com/golang/mock) version found in `api/mocks/realm_api.go`

Run:

```
go run github.com/golang/mock/mockgen -source ./api/realm_client.go -destination ./api/mocks/realm_client.go
```
