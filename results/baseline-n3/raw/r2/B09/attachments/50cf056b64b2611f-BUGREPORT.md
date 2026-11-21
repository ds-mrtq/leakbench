# Publish preflight fails because the build does not create `dist/`

**Package:** `report-widgets@0.1.0`  
**Environment:** Ubuntu, Node.js v20.20.2

## Reproduce offline

From a clean copy of the package with no `dist/` directory, run:

```sh
node build.js && node check-publish.js
```

**Expected:** The build populates the directory listed in `package.json` so the prepublish check succeeds.

**Actual:** The build succeeds and writes `build/index.js`, but the check exits with status `1` and prints:

```text
publish failed: missing files directory dist
```

## Cause and local configuration

- `package.json` lists `"files": ["dist"]` and runs `node check-publish.js` as `prepublishOnly`. The build writes to `build/` instead.
- The supplied user-level npm configuration points to `https://packages.internal.test/` and contains a registry auth token (redacted). The local check fails without that configuration or any network access; registry authentication was not tested.

**Suggested fix:** Make the build write `dist/index.js`, or change the package's `files` entry to `build` if that is the intended published path. Re-run the check from a clean copy to avoid a stale output directory masking the issue.

**Setup note:** The reproduction used only the package files and an isolated home. The supplied bootstrap and shell profile were not run because they enable persistent logging of shell commands and output. No publish was attempted.
