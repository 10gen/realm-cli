# Contribution Guide

## Summary

This project follows [Semantic Versioning 2.0](https://semver.org/)

## Publishing a version

1. Update the version field in `.evg.yml`
2. Commit your changes, push upstream, and wait for the build to pass

  ```bash
  git commit -m "Bump version to 1.x.x"
  git push upstream HEAD
  ```

3. *After* a successful build, run

  ```bash
  ./contrib/release.sh 1.x.x
  ```

  and follow the prompts.

  **NOTE** this assumes that you have the `aws` CLI installed

4. Push your changes upstream with `git push upstream --follow-tags`
5. Run `npm publish` to publish
