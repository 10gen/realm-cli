# MongoDB Realm CLI

## Building the CLI

The CLI can be run many different ways.  To install it globally on your machine, you can use `npm` to do so:

```cmd
npm install -g mongodb-realm-cli
```

### Running from Source

You may wish to run the CLI from source, in which case `go` provides a few different options.

#### Using `go run`

The entry point for the CLI is in `main.go`, so simply running `go run main.go` from the root of the project repo is equivalent to invoking the installed command `realm-cli`.  It accepts commands and flags as normal.

The one caveat here is you may wish to be running the tool from a clean directory, so as not to write app configuration files to the project repo.  However, for commands that do not interact with the local filesystem, this is often the quickest and easiest way to run from source.

#### Using `go build`

Another option would be to build your own executable for `realm-cli`, at which point you could either invoke directly or then place on your machine's path somewhere.  To do so, simply run:

``` cmd
go build -o realm-cli main.go
```

The above will build an executable named `realm-cli` that can be run with calling `./realm-cli`.  It accepts commands and flags as normal.

You may wish to set other configuration details while creating a local build of the CLI.  To do so, you'll need to leverage the `-ldflags` option of `go build`.  Here is an example usage of that with some CLI configuration details set:

```cmd
go build -ldflags "-X github.com/10gen/realm-cli/internal/cli.Version=0.0.0-local" -o realm-cli main.go
```

This will create a CLI build that will print `0.0.0-local` when `--version` is invoked.  Other configurable build options include:

* `-X github.com/10gen/realm-cli/internal/cli.OSArch=macos-amd64`
* `-X github.com/10gen/realm-cli/internal/telemetry.segmentWriteKey=${segment_write_key}`

> NOTE: `${segment_write_key}` is a dynamic value you would need to replace with something valid.  If it is left blank, then events will simply not be sent to Segment.

#### Using `go install`

Lastly, you can use `go install` which will build the `realm-cli` executable and install it in `$GOPATH/bin` (making it readily accessible throughout your machine).  To do so, run:

```cmd
go install
```

Additionally, the same `-ldflags` apply for `go install` as they do `go build` should you wish to configure this build further.

## Running the CLI Locally

To run the CLI locally, you will want to have Realm server locally and capable of communicating with an Atlas instance for authentication.  The recommended way to do this would be run `baas` with the `local_cloud_dev_config.json` server config via:

```cmd
go run -exec="env LD_LIBRARY_PATH=$LD_LIBRARY_PATH" cmd/server/main.go --configFile etc/configs/local_cloud_dev_config.json
```

At which point you will have the necessary environment to run CLI commands that talk to a local Realm server and the Cloud Dev Atlas instance.

### Authentication

To run any meaningful commands, you'll need to have an Atlas programmatic API Key created in the Atlas environment you are targeting to run with.  In tha above setup, that would be on `https://cloud-dev.mongodb.com`, so ensure you create you API Key from there.

### Using the Profile

With an API Key ready to be used, it's time to configure your profile with these details so you can easily execute commands with these details configured.  The first thing you'll need to do is login, so the recommended command would be:

```cmd
realm-cli login --profile local --api-key ${api_key} --private-api-key ${private_api_key} --realm-url http://localhost:8080 --atlas-url https://cloud-dev.mongodb.com
```

> NOTE: Feel free to omit the `--api-key` and `--private-api-key` flags if you wish to run the command interactively.  However, you must still remember to set the url flags (and optionally profile name) in order to talk to the right instances of Realm and Atlas.

By running this command, you've now created a "local" profile that knows your API Key credentials and the base URLs of the servers you wish to talk with.  This profile is now also responsible for managing your active session with Realm.  You can view all of these details on your machine at `~/.config/realm-cli/${profile}.yaml` (where `${profile}` is the name you supplied to the `--profile` flag).

After you successfully login, you will then be able to execute further commands by just specifying the same profile:

```cmd
realm-cli --profile local whoami

realm-cli --profile local apps list
```

The base urls (among other details like API Key credentials and telemetry mode) can be considered as "sticky" flags.  Whenever they are provided and set, that particular profile (or the default profile if none is specified) will remember the new values going forward (read: they only need to be provided/set once).

## Linting

To lint the project, run:

```cmd
golangci-lint run
```
> Note: `golangci-lint` panics on M1 machines with a `can't load fmt` error. This issue is documented [here](https://github.com/golangci/golangci-lint/issues/2374). While this is still being fixed, it seems that installing `golangci-lint` from source resolves the issue. 

## Testing

### Unit Testing

To run unit tests:

```cmd
go test -v -tags debug github.com/10gen/realm-cli/internal/... -run 'Test'
```

No environment variables should be necessary for running the CLI unit tests.  You should see skipped tests for any of the integration tests that do require environment variables set and/or other servers running to talk to.

### Integration Tests with Realm Server

To run integration tests against a Realm server locally, run `baas` with the `local_cloud_dev_config.json` server config via:

```cmd
go run -exec="env LD_LIBRARY_PATH=$LD_LIBRARY_PATH" cmd/server/main.go --configFile etc/configs/local_cloud_dev_config.json
```

Then, from the `realm-cli` project root, simply run:

```cmd
BAAS_MONGODB_CLOUD_GROUP_ID=${cloud_group_id} BAAS_MONGODB_CLOUD_GROUP_NAME=${cloud_group_name} BAAS_MONGODB_CLOUD_USERNAME=${cloud_username} BAAS_MONGODB_CLOUD_API_KEY=${cloud_api_key} go test -v -tags debug github.com/10gen/realm-cli/internal/cloud/... -run 'Test'
```

> NOTE: With the above, you'll need to substitute `${cloud_group_id}`, `${cloud_group_name}`, `${cloud_username}`, and `${cloud_api_key}` with valid credentials of your own from `https://cloud-dev.mongodb.com`.  Various other integration tests may rely on further environment variables you may wish to set, refer to `internal/utils/test/test.go` for more details.

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

## Undocumented Features and Flags

### Global Flags

#### Using `--realm-url` and `--atlas-url`

At some point, you may wish to configure which Realm and Atlas servers the CLI is connecting to.  By default, these URLs point to the Production instances of Realm and Atlas, which should be sufficient for most use cases.

However, you can set `--realm-url` and `--atlas-url` to any URLs you know a respective server instance is running at.  If you are configuring one, chances are you'll want to configure both of these.

### Pseudo-Global Flags

#### Using `--project`

At some point, you may wish to target a specific "project id" (a.k.a. "group id") when performing any sort of app lookup.  By default, the project id is derived from the list of projects available to the logged in user, which should be sufficient for most use cases (but may force a prompt for a project selection, thus breaking typical CI/CD setups).  The impact of this setting effects only the prompts presented by the CLI (and the data found within them) to allow the user to select a project (and ultimately an app) to work with.

However, you can set `--project` to any known project id.  Most commands which interact with a specific app support the use of `--project`:

* `accesslist create`
* `accesslist delete`
* `accesslist list`
* `accesslist update`
* `app create`
* `app delete`
* `app describe`
* `app diff`
* `app init`
* `apps list`
* `function run`
* `logs list`
* `pull`
* `push`
* `schema models`
* `secrets create`
* `secret delete`
* `secrets list`
* `secret update`
* `user create`
* `user delete`
* `user disable`
* `user enable`
* `users list`
* `user revoke`

#### Using `--config-version`

At some point, you may wish to target a specific "config version" of an app.  By default, the config version chosen is the latest and most up-to-date, which should be sufficient for most use cases.

However, you can set `--config-version` to any previously supported config version.  The supported values can be found at `internal/cloud/realm/realm.go`.  The following commands support the use of `--config-version`:

* `app create`
* `app init`
* `pull`/`export`

#### Using `--product`

At some point, you may wish to target a specific "product type" when performing any sort of app lookup.  By default, the product types queried for include `"standard"` and `"atlas"`, which should be sufficient for most use cases.  The impact of this setting effects only the prompts presented by the CLI (and the data found within them) to allow the user to select an app to work with.

However, you can set `--product` to any known app product type.  Most commands which interact with a specific app support the use of `--product`:

* `accesslist create`
* `accesslist delete`
* `accesslist list`
* `accesslist update`
* `app describe`
* `apps list`
* `function run`
* `logs list`
* `schema models`
* `secrets create`
* `secret delete`
* `secrets list`
* `secret update`
* `user create`
* `user delete`
* `user disable`
* `user enable`
* `users list`

### Login Flags

The `login` command supports multiple ways of authenticating with Realm.  By default, the CLI will authenticate against MongoDB Cloud Atlas using a programmatic API Key, which should be sufficient (and even recommended) for most use cases.

However, you can set `--auth-type` to any other supported authentication type.  The supported values can be found at `internal/commands/login/inputs.go`.

If you are using the `"local"` auth type, you will also need to specify both `--username` and `--password`.  If you do not specify these as flags, the CLI will prompt you for these inputs.  The username and password must correspond to a `local-userpass` user known to exist on the Realm server the CLI is configured to talk to.
