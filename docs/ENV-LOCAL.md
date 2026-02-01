# .env.local Standard (Agent Auth)

The `.env.local` file is required for agent authentication and is not committed.

Required keys (v0):

- GITHUB_TOKEN

Optional keys:

- GITHUB_REPO
- GITHUB_BASE_BRANCH

Agent usage:

1. `source .env.local`
2. `echo "$GITHUB_TOKEN" | gh auth login --with-token`
3. `gh auth status`

Security note:

- Tokens are repo-scoped and rotated; do not paste tokens into chat or commits.
