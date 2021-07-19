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

1. Update the version field in `.evg.yml` and commit your changes
  ```bash
  git commit -m "Bump version to 2.x.x"
  ```

2. Push to your fork and create a PR
  ```bash
  git push origin HEAD
  ```

3. After merging your PR, wait for the build to pass

4. *After* a successful build, run

  ```bash
  ./contrib/release.js 2.x.x
  ```

  and follow the prompts. This will update local files and commit the changes for the release.

  **NOTE** this assumes that you have the `aws` CLI installed

5. Push the commit to your fork and create another PR
6. After merging, push the new tag upstream
```bash
git push upstream --tags
```

7. Run `npm publish` to publish
