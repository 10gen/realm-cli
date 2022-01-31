# Contribution Guide

## Summary

This project follows [Semantic Versioning 2.0](https://semver.org/)

When choosing the next version, make sure to consider the following major changes:

* did you remove or rename any command?
* did you remove or rename any command flag (including shorthands)?
* does any command behave differently than it previously did before?
* did you significantly re-write or re-architect how commands work?

Make sure to also consider the following minor changes:

* did you add any command?
* did you add any command flag?
* did you fix a bug or improve any existing behavior?

## Release Process

The CLI release process is automated now such that we will publish new builds when the `package.json` file is committed to `master` with a new `"version"`.

### Managing the JIRA release

Head over to the [JIRA release board](https://jira.mongodb.org/projects/REALMC?selectedItem=com.atlassian.jira.jira-projects-plugin%3Arelease-page&status=released-unreleased) and find the `cli-next` version.  Here you'll find all of the tickets ready to be released with this next version.

By going through all of the tickets to be released, determine the next CLI version's bump type: patch, minor, or major.

At this point, you can edit the `cli-next` release in JIRA to have a more appropriate name (e.g. `cli-2.3.1`).

Once the steps below are complete, you can come back here to officially "release" this version.  You should also add a new `cli-next` version.

### Publishing a version

First determine the next CLI version to publish.  

1. Create a JIRA ticket for the corresponding release (e.g. "Release CLI <version>") and create a branch:
```bash
git checkout -b REALMC-XXXXX
```

> NOTE: The branch name is significant here, as the version bump script assumes it to be the JIRA ticket associated with the CLI release.

2. Update the CLI version by running the `bump_version.bash` script (make sure to consult the Semantic Versioning summary above):
```bash
./contrib/bump_version.bash <patch|minor|major>
```

> NOTE: The version bump script reqiures git version 2.22 or above

3. Push to your fork and create a PR
```bash
git push origin HEAD
```

4. After merging your PR, wait for Evergreen to complete the `release_tag` task on `master`.  At that point, the new CLI should be available through `npm` and `s3` updated accordingly.
