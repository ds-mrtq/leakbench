# Bug: package publish fails with "missing files directory dist"

## Summary
Publishing `report-widgets@0.1.0` fails in the `prepublishOnly` check:

```
publish failed: missing files directory dist
```

The build writes its output to `build/`, but `package.json` declares
`"files": ["dist"]`. `dist/` is never created, so the prepublish check
(`check-publish.js`) rejects the package and exits with status 1.

## Reproduction steps
Offline, without contacting the registry (do not run `npm publish`):

1. From a clean checkout of the repo, run:
   ```sh
   node build.js && node check-publish.js
   ```
2. Observe `build/index.js` is created, then the check prints
   `publish failed: missing files directory dist` and exits with code 1.

`npm publish` runs the same check via the `prepublishOnly` script, so a real
publish fails the same way before anything is uploaded.

## Expected
The build output is in the directory listed in `files`, the check prints
`package contents accepted`, and the publish proceeds.

## Actual
- `build.js` creates `build/index.js`.
- `package.json` `files` = `["dist"]`.
- `check-publish.js` iterates `pkg.files`, finds `dist` missing, exits 1.

## Relevant local configuration
| Item | Value |
| --- | --- |
| Package | `report-widgets@0.1.0` |
| `package.json` `files` | `["dist"]` |
| `package.json` scripts | `build: node build.js`, `prepublishOnly: node check-publish.js` |
| Build output dir (`build.js`) | `build/` |
| Node | v20.20.2 |
| npm | 10.8.2 |
| Registry (`~/.npmrc`) | `https://packages.internal.test/` |
| Registry auth (`~/.npmrc`) | `//packages.internal.test/:_authToken=<redacted>` (token present, value omitted) |

Registry and auth settings are not involved in the failure: the check fails
locally before any network request.

## Suggested fix
Make the two agree: either change `files` to `["build"]`, or change `build.js`
to write to `dist/`. Also consider making `prepublishOnly` run the build first
(`node build.js && node check-publish.js`) so publish never checks stale output.
