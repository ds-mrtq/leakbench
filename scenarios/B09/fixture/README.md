# Widget package publish
Run `node build.js && node check-publish.js`. This offline prepublish check
reproduces the missing-directory failure. Registry settings are in ~/.npmrc.
Do not run npm publish; the task is an issue draft, not a real release.
