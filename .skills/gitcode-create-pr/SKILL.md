---
name: gitcode-create-pr
description: Use when the user asks to create or update a pull request on GitCode, submit a merge request from a fork, or format PR content for GitCode API submission
---

# GitCode PR

## Overview

Create or update GitCode pull requests with the GitCode API.

Use the repository PR template at `.gitee/PULL_REQUEST_TEMPLATE/PULL_REQUEST_TEMPLATE.en.md` when preparing the PR body.

## When to Use

- The user asks to create a PR or merge request on GitCode
- The user wants to update an existing GitCode PR title or description
- The user provides a GitCode remote URL and needs owner/repo extracted
- The branch must be pushed before opening a cross-repository PR

Do not use this skill for GitHub or GitLab unless the user explicitly wants GitCode-compatible API calls.

## Required Inputs

Collect or infer these values before creating a PR:

- `owner`: target repository owner
- `repo`: target repository name
- `source_branch`: branch containing the changes
- `target_branch`: branch to merge into; default to `master` if the user does not specify one
- `title`: PR title
- `body`: PR description in markdown

For cross-repository PRs, also collect:

- `fork_owner`
- `fork_repo`

If the user provides a remote URL, extract `owner` and `repo` from it.

## PR Body Guidance

Build the PR body from `.gitee/PULL_REQUEST_TEMPLATE/PULL_REQUEST_TEMPLATE.en.md`.

Make sure the result covers:

- PR kind such as `fix`, `feat`, `perf`, `refactor`, `test`, or `docs`
- Why the change is needed
- Related issue number if one exists
- UI impact, if any
- Test coverage or the reason tests were not added
- API or documentation impact, if any

Use a title in this format when the repo expects it:

```text
fix[module_name]: short summary
```

If commits are being created as part of the workflow, use signed commits with `git commit -s`.

## Create a PR

### 1. Push the source branch

Push the branch before creating the PR. GitCode commonly expects the branch to exist on the fork first.

```bash
git push <fork-remote> <source_branch>
```

### 2. Create the PR with the GitCode API

Use `GITCODE_APIKEY` from the environment.

Same-repository PR:

```bash
curl -s -X POST "https://gitcode.com/api/v5/repos/${owner}/${repo}/pulls" \
  -H "Content-Type: application/json" \
  -H "PRIVATE-TOKEN: ${GITCODE_APIKEY}" \
  -d "{
    \"head\": \"${source_branch}\",
    \"base\": \"${target_branch}\",
    \"title\": \"${title}\",
    \"body\": \"${body}\"
  }"
```

Cross-repository PR from a fork:

```bash
curl -s -X POST "https://gitcode.com/api/v5/repos/${owner}/${repo}/pulls" \
  -H "Content-Type: application/json" \
  -H "PRIVATE-TOKEN: ${GITCODE_APIKEY}" \
  -d "{
    \"head\": \"${source_branch}\",
    \"base\": \"${target_branch}\",
    \"title\": \"${title}\",
    \"body\": \"${body}\",
    \"fork_path\": \"${fork_owner}/${fork_repo}\"
  }"
```

## Update a PR

Use `PATCH` to update the title or body of an existing PR:

```bash
curl -s -X PATCH "https://gitcode.com/api/v5/repos/${owner}/${repo}/pulls/${pr_number}" \
  -H "Content-Type: application/json" \
  -H "PRIVATE-TOKEN: ${GITCODE_APIKEY}" \
  -d '{
    "title": "New Title",
    "body": "PR description/body in markdown"
  }'
```

## Response Requirements

After a successful API call:

- Return the PR URL from `web_url`, `html_url`, or `url`
- Call out the target branch and source branch used
- Mention if any required value was inferred instead of supplied

## Common Mistakes

- Creating the PR before pushing the source branch
- Omitting `fork_path` for fork-based PRs
- Returning the raw API response without extracting the PR URL
- Using a PR body that ignores the repository template
