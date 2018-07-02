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
gometalinter --exclude=vendor --vendor --config=.linter.config ./...
```

## Testing

Run all tests:

```go
go test -v $(go list github.com/10gen/stitch-cli/...)
```
