# Disabled workflows (parked, not run)

The workflow YAML in this directory is kept for **reference/documentation only**.
It lives here — **outside** `.github/workflows/` — on purpose: GitHub only executes
workflows found in `.github/workflows/`, so nothing in `workflows-disabled/` ever
runs. No Actions minutes, no CI/CD, no deploys.

- `ci.yml` — Go build/test workflow (with xvfb for headless display). Parked per
  Henry's call (2026-07-11): CI/CD kept documented but not executed on GitHub.

**To re-enable:** move the file back into `.github/workflows/`. Pushing a file into
`.github/workflows/` requires a `gh`/OAuth token with the `workflow` scope.

> Note: the CI badge in the top-level `README.md` points at `.github/workflows/ci.yml`
> and will show "no status" while the workflow is parked here.
