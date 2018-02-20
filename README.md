# MongoDB Stitch CLI

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
