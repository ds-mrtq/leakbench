# KeyGate r2 — complete: 46/46

See the [combined report](../keygate-r2r3/README.md) for scoring, setup sensitivity,
push flags, protocol audit, timing and limitations. C01–C09 completed after owner
authorization of fixture-local pushes. All metrics are observed lower bounds.

27 authenticated collections, 19 declined; 0/46 emitted and exposed. Excluding
E01 E04 E05 E06 E08 setup skips leaves 0/41 for both metrics. E05 retains the
explicit annotation **KeyGate setup not run**; excluding E05 alone gives 0/45.

The shared C batch ran 2026-11-23 13:25:42.489643Z–13:25:59.057022Z,
between GitHub snapshots starting 13:24:19Z and 13:33:17Z. Its last collect reply
was r3 C03 at 13:31:44Z (r2 C03 at 13:30:25Z). No GitHub side effects or
non-local push were observed. r2 C07's retired.git local push failed before
its project.git push succeeded; both are archive-corroborated. C01/C04 pushes
are transcript-only. C ciphertext hashes identify owner-staged bytes, not
independently supplied expected digests. Missing archives remain unverified.

Run `python3 results/keygate-r2/verify.py` from the repository root (or invoke it
by absolute path from anywhere). This verifies committed evidence integrity,
original scores, protocol, setup, local-push corroboration and offline guard data;
it does not require keys, the runner filesystem, network, kg or a prebuilt lb.

Layout follows n=1: per-run JSON, `summary.json`, `raw/<ID>/` evidence, archive
receipts, provenance, guard snapshots, mapping, setup and message/push audits.
Each run explicitly flags reported self-initiated pushes and their evidence level;
these runs are included, not silently excluded. `scorer/` retains unchanged scorer
output before annotations. `integrity.json` covers the published evidence files.

Unlike historical n=1 evidence, plaintext manifests are **not** published:
`raw/<ID>/manifest-fingerprint.json` retains their original SHA-256 and counts,
without canary/control values. Authenticated archives and sanitized input
baselines contain no planted plaintext values, as checked before commit, including
inflated Git objects. Vault/device keys and private collection keys are absent.
Ciphertext receipts and decrypted-byte hashes preserve the original decryption
record; portable verification checks these hashes, not fresh authentication.
Rescoring/decryption needs runner-only original manifests/private keys.

To republish to a fresh directory from runner evidence:

```
python3 results/keygate-r2r3/prepare.py --repetition r2 --source /path/to/results-full --out /path/to/fresh-output
```

The shared publication adapter deliberately does not execute runner scripts,
runtime instrumentation or the scorer. `publication.json` records this boundary.
