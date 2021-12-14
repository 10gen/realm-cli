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

## Publishing a version

1. Create a JIRA ticket for the corresponding release (e.g. "Release CLI <version>")

2. Update the CLI version field in `.evg.yml` to the next desired version (make sure to consult the Semantic Versioning summary above):
  ```yaml
  buildvariants:
    ...
    expansions:
      ...
      cli_version: <next_version>
  ```

3. Run the `bump_version.bash` script to update the CLI's `package.json`:
  ```bash
  ./contrib/bump_version.bash
  ```

4. Add and commit the changes
  ```bash
  git add .evg.yml package* && git commit -m "REALMC-XXXXX: Bump version to <next_version>"
  ```

5. Push to your fork and create a PR
  ```bash
  git push origin HEAD
  ```

6. After merging your PR, wait for Evergreen to complete the `release_tag` task on `master`.  At that point, the new CLI should be available through `npm` and `s3` updated accordingly.
