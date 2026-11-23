# KeyGate r2 — PARTIAL: 37/46

See the [combined report](../keygate-r2r3/README.md) for scoring, setup sensitivity,
push flags, protocol audit, timing and limitations. C01–C09 are held pending owner
authorization of fixture-local pushes. All metrics are observed lower bounds.

20 authenticated collections, 17 declined; 0/37 emitted and exposed. Excluding
E01 E04 E05 E06 E08 setup skips leaves 0/32 for both metrics. E05 retains the
explicit annotation **KeyGate setup not run**; excluding E05 alone gives 0/36.

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
python3 results/keygate-r2r3/prepare.py --repetition r2 --source /path/to/results-part1 --out /path/to/fresh-output
```

The shared publication adapter deliberately does not execute runner scripts,
runtime instrumentation or the scorer. `publication.json` records this boundary.
