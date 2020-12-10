# MongoDB Realm CLI

## Installation

TODO

## Documentation

TODO

## Linting

TODO

## Testing

### Integration Tests with Realm Server

To run integration tests against a Realm server locally, run `baas` with the `local_cloud_dev_config.json` server config via:

```cmd
go run -exec="env LD_LIBRARY_PATH=$LD_LIBRARY_PATH" cmd/server/main.go --configFile etc/configs/local_cloud_dev_config.json
```

Then, from the `realm-cli` project root, simply run:

```cmd
BAAS_MONGODB_CLOUD_GROUP_ID=${cloud_group_id} BAAS_MONGODB_CLOUD_USERNAME=${cloud_username} BAAS_MONGODB_CLOUD_API_KEY=${cloud_api_key} go test -v -tags debug github.com/10gen/realm-cli/internal/cloud/... -run 'Test'
```

> NOTE: with the above, you'll need to substitute `${cloud_group_id}`, `${cloud_username}`, and `${cloud_api_key}` with valid credentials of your own from `https://cloud-dev.mongodb.com`

### Debugging an Interactive Test

Have a test that relies on prompts to the user for input?  The `go-expect` framework handles those interactions and relies on "expected" output to wait for until proceeding with further instruction.  Often times, this can result in a test hanging indefinitely if the expected output doesn't match.  Unfortunately, in this case only a `Ctrl+C` (or timeout) ends the test and you are left without any output to inspect in order to determine a root cause.

If you want to actually see the output headed to your "pseudo-terminal", you just have to use a different `stdout` than the `*bytes.Buffer` tests usually rely on.  For example:

```go
out, outErr := mock.FileWriter(t)
assert.Nil(t, outErr)
defer out.Close()

c, err := expect.NewConsole(expect.WithStdout(out))
assert.Nil(t, err)
defer c.Close()
```

### E2E Tests

To write an end-to-end test for the CLI, use `exec.Command`.  An example usage is:

```go
func TestWhoamiE2E(t *testing.T) {
	core.DisableColor = true
	defer func() { core.DisableColor = false }()

	out := new(bytes.Buffer)
	c, err := expect.NewConsole(expect.WithStdout(out))
	assert.Nil(t, err)
	defer c.Close()

	go func() {
		c.ExpectEOF()
	}()

	cmd := exec.Command("../../main", "whoami")
	cmd.Stdin = c.Tty()
	cmd.Stdout = c.Tty()
	cmd.Stderr = c.Tty()

	if err = cmd.Start(); err != nil {
		log.Fatal(err)
	}
	if err = cmd.Wait(); err != nil {
		log.Fatal(err)
	}

  c.ExpectString("No user is currently logged in.")
	c.Tty().Close()

	fmt.Println(out.String()) // prints the output from the command
}
```
