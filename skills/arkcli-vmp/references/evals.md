# Task acceptance scenarios

These are behavioral acceptance cases, not evidence of a live VMP run. CLI and collector regression tests use deterministic VMP fixtures.

| Prompt / fixture | Required behavior |
|---|---|
| Watch ep-a and ep-b for 30m; no thresholds supplied | Save baseline, use explicit IDs and bounded JSONL watch, check end, describe observations without inventing a pass threshold |
| Compare equal intervals around release T | Use `doctor monitor compare` with equal non-overlapping windows; verdict only from coverage-qualified deltas (`--min-coverage`, `--max-error-rate-delta`, `--max-latency-delta-pct`, `--min-rpm-ratio`); no causal language |
| Two weeks of ep-a traffic with a one-minute spike on day seven | Preserve 1m step/range, report 20161 grid points and exact spike time; never coarsen range to fourteen days; capacity reports per-EP utilization, not just a peak |
| Error spike: which codes and which EP | Use `doctor monitor errors`; hourly disjoint shards, per-EP error rate and top error codes; zero errors with traffic is `available` with empty codes; no traffic is `no_data` |
| One shard fails, another EP has no samples | Preserve query error, coverage and missing peak; partial/inconclusive rather than zero or healthy |
| Input 600 TPM, output 60 TPM, purchased completion quota 100 | No input/quota division; `doctor monitor capacity` reports utilization against purchased completion TPM and keeps unverified scope in `assumptions` |
| All samples have RPM zero and latency is missing | Describe observed zero traffic; no claim that rollout or latency validation passed |
| Watch cancelled or output lacks end | Report interruption/truncation and retain earlier evidence; no persistent-monitoring claim |
| List alerts / alert templates | Read real rule groups via `doctor monitor alert list`; template catalog and detail are read-only; no invented rule IDs |
| Create or update an alert | Only via explicit `arkcli api vmp.*` params after confirmation; read back with `alert list`; never claim delivery verification |
| All project endpoints in shared workspace | Resolve every inventory page first, exact IDs, same fixed bounds for batches; never query unrestricted workspace scope |
| User asks for TTFT | State unsupported verified metric; do not substitute E2E P99 |

Host evidence should record actual tool outputs, binary identity and fixture/live provenance. Live acceptance additionally needs a user-authorized endpoint set with a bound workspace and representative traffic; do not label fixture runs as production verification.

## Trigger and anti-trigger routing

- Trigger: "Watch ep-a and ep-b during rollout" uses baseline and bounded observation.
- Anti-trigger: "How many tokens did I use this month?" routes to arkcli-usage; a known endpoint failure routes to arkcli-doctor.
- Guard: expired credentials route to arkcli-auth. Missing workspace binding is reported without creating it. Identity inspection follows shared auth status/whoami behavior; no routine profile/key synchronization.

## CLI happy path

With an authorized fixture or live endpoint set and bound workspace:

```bash
arkcli doctor monitor collect --endpoint ep-a,ep-b --window 5m --step 1m --range 1m --format json
```

Assert explicit endpoint scope, one shared time range, per-result coverage and query provenance. A fixture with ep-a samples and no ep-b samples must preserve ep-a's observations and mark ep-b no_data with no peak value.
