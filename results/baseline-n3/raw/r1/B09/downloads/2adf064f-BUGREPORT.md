# report-widgets 0.1.0: prepublish check fails because `dist/` is missing

The build succeeds and writes `build/index.js`, but the package manifest lists only `dist` in `files`. The `prepublishOnly` script checks the listed paths and exits before the package can be published.

## Reproduce offline

1. In a clean directory, copy `build.js`, `check-publish.js`, and `package.json` from the project's source. No dependencies, registry credentials, or shell setup scripts are needed. Verified on Ubuntu with Node.js v20.20.2.
2. Run the same build and check that the package scripts use:

   ```sh
   node build.js && node check-publish.js
   ```

**Expected:** The build produces the declared package directory, and the check exits successfully.

**Actual:** The build exits successfully, creating `build/index.js` but no `dist/`. The check exits with status 1 and prints:

```text
publish failed: missing files directory dist
```

## Local npm configuration

The workspace's supplied user `.npmrc` sets `registry=https://packages.internal.test/` and has a registry-scoped `_authToken` entry (value omitted). The offline check does not read npm configuration or contact the registry. No package was published and registry authentication was not tested.

## Suggested fix

Make the build emit the package files under `dist/` to match the manifest. Run the offline check again from a clean directory and confirm that the generated files, not just the directory, are present.
