---
name: gitcode-api
description: Interact with GitCode API - repos, issues, PRs, users, organizations
---

# GitCode API Skill

## Base URL
```
https://api.gitcode.com/api/v5
```

## Authentication
| Method | Header/Param |
|--------|--------------|
| Bearer Token | `Authorization: Bearer {token}` |
| Private Token | `PRIVATE-TOKEN: {token}` |
| Query Param | `?access_token={token}` |

Tokens: https://gitcode.com/setting/token-classic

## Rate Limits
- 50 requests/minute
- 4000 requests/hour
- 429 = rate limited

## Core Endpoints

### User
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/user` | Get authenticated user |
| PATCH | `/user` | Update authenticated user profile |
| GET | `/users/{username}` | Get user info |
| GET | `/users/{username}/events` | Get user activity |
| GET | `/users/{username}/repos` | Get user's public repos |
| GET | `/users/{username}/merge_requests` | Get user's MRs |
| GET | `/user/repos` | List user's repos |
| POST | `/user/repos` | Create user repo |
| GET | `/user/merge_requests` | List authenticated user's MRs |
| GET | `/user/keys` | List user's SSH keys |
| POST | `/user/keys` | Add SSH key |
| DELETE | `/user/keys/{id}` | Delete SSH key |
| GET | `/user/keys/{id}` | Get SSH key |
| GET | `/user/namespace` | Get user's namespace |
| GET | `/emails` | Get user's emails |

### Repositories
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/repos/{owner}/{repo}` | Get repo info |
| GET | `/repos/{owner}/{repo}/branches` | List branches |
| GET | `/repos/{owner}/{repo}/commits` | List commits |
| GET | `/repos/{owner}/{repo}/tags` | List tags |
| POST | `/repos/{owner}/{repo}/tags` | Create tag |
| DELETE | `/repos/{owner}/{repo}/tags/{tag_name}` | Delete tag |
| GET | `/repos/{owner}/{repo}/git/trees/{sha}` | Get directory tree |
| GET | `/repos/{owner}/{repo}/contents/{path}` | Get file content |
| GET | `/repos/{owner}/{repo}/forks` | List forks |
| GET | `/repos/{owner}/{repo}/hooks` | List webhooks |
| POST | `/repos/{owner}/{repo}/hooks/{id}/tests` | Test webhook |
| GET | `/repos/{owner}/{repo}/releases` | List releases |
| POST | `/repos/{owner}/{repo}/releases` | Create release |
| GET | `/repos/{owner}/{repo}/releases/{tag}` | Get release by tag |
| GET | `/repos/{owner}/{repo}/releases/latest` | Get latest release |
| GET | `/repos/{owner}/{repo}/releases/tags/{tag}` | Get release by tag name |
| GET | `/repos/{owner}/{repo}/milestones` | List milestones |
| POST | `/repos/{owner}/{repo}/milestones` | Create milestone |
| GET | `/repos/{owner}/{repo}/milestones/{number}` | Get milestone |
| PATCH | `/repos/{owner}/{repo}/milestones/{number}` | Update milestone |
| DELETE | `/repos/{owner}/{repo}/milestones/{number}` | Delete milestone |
| PUT | `/repos/{owner}/{repo}/project/labels` | Update labels |
| GET | `/repos/{owner}/{repo}/protected-tags` | List protected tags |
| POST | `/repos/{owner}/{repo}/protected-tags` | Create protected tag |
| PUT | `/repos/{owner}/{repo}/protected-tags/{tag_name}` | Update protected tag |
| DELETE | `/repos/{owner}/{repo}/protected-tags/{tag_name}` | Delete protected tag |
| PUT | `/repos/{owner}/{repo}/collaborators/{username}` | Add/update collaborator |
| DELETE | `/repos/{owner}/{repo}/collaborators/{username}` | Remove collaborator |
| GET | `/repos/{owner}/{repo}/collaborators` | List collaborators |
| GET | `/repos/{owner}/{repo}/collaborators/{username}/permission` | Get collaborator permission |
| GET | `/repos/{owner}/{repo}/notifications` | List notifications |
| PUT | `/repos/{owner}/{repo}/notifications` | Mark notifications as read |

### Issues
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/repos/{owner}/{repo}/issues` | List issues |
| POST | `/repos/{owner}/issues` | Create issue |
| GET | `/repos/{owner}/{repo}/issues/{number}` | Get issue |
| PATCH | `/repos/{owner}/issues/{number}` | Update issue |
| PUT | `/repos/{owner}/issues/{number}/related-branches` | Set issue related branches |
| GET | `/repos/{owner}/{repo}/issues/{number}/comments` | List issue comments |
| POST | `/repos/{owner}/{repo}/issues/{number}/comments` | Add issue comment |
| PATCH | `/repos/{owner}/{repo}/issues/comments/{id}` | Update comment |
| DELETE | `/repos/{owner}/{repo}/issues/comments/{id}` | Delete comment |
| POST | `/repos/{owner}/{repo}/issues/{number}/labels` | Add label to issue |
| DELETE | `/repos/{owner}/{repo}/issues/{number}/labels/{name}` | Remove label from issue |

### Pull Requests (MRs)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/repos/{owner}/{repo}/pulls` | List MRs |
| POST | `/repos/{owner}/{repo}/pulls` | Create MR |
| GET | `/repos/{owner}/{repo}/pulls/{number}` | Get MR |

### Organizations
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/orgs/{org}` | Get org info |
| PATCH | `/orgs/{org}` | Update org profile |
| GET | `/orgs/{org}/repos` | List org repos |
| POST | `/orgs/{org}/repos` | Create org repo |
| GET | `/orgs/{org}/members` | List org members |
| GET | `/orgs/{org}/members/{username}` | Get org member |
| POST | `/orgs/{org}/memberships/{username}` | Invite member |
| DELETE | `/orgs/{org}/memberships/{username}` | Remove member |
| GET | `/orgs/{org}/issue-extend-settings` | Get issue config |

### Users & Orgs
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/users/{username}/starred` | List starred repos |
| GET | `/users/{username}/subscriptions` | List watched repos |
| GET | `/user/starred` | List authenticated user's starred repos |
| GET | `/user/subscriptions` | List authenticated user's watched repos |
| GET | `/user/namespaces` | List user's namespaces |
| GET | `/org/{owner}/kanban/list` | Get kanban board |

### Enterprise
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/enterprises/{enterprise}/members` | List enterprise members |
| GET | `/enterprises/{enterprise}/members/{username}` | Get enterprise member |
| PUT | `/enterprises/{enterprise}/members/{username}` | Update member permissions |
| GET | `/enterprises/{enterprise}/labels` | List enterprise labels |

### Search
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/search/users?q={keyword}` | Search users |
| GET | `/search/repositories?q={keyword}` | Search repositories |
| GET | `/search/issues?q={keyword}` | Search issues |

### Other
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/chat/completions` | AI hub completions |
| GET | `/oauth` | OAuth2.0 authentication |

## Usage

```bash
# Base URL
BASE_URL="https://api.gitcode.com/api/v5"

# List branches
curl -H "PRIVATE-TOKEN: $GITCODE_TOKEN" \
  "${BASE_URL}/repos/{owner}/{repo}/branches"

# Create issue
curl -X POST -H "PRIVATE-TOKEN: $GITCODE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Bug","body":"Description"}' \
  "${BASE_URL}/repos/{owner}/issues"

# Search users
curl "${BASE_URL}/search/users?q={query}"
```

## Common Request/Response Examples

### Search Users
**Request:**
```bash
curl "${BASE_URL}/search/users?q=keyword"
```
**Response:** 200
```json
{
  "data": [
    {
      "id": 123,
      "username": "user",
      "name": "User Name",
      "avatar_url": "https://...",
      "web_url": "https://gitcode.com/user"
    }
  ],
  "total": 1
}
```

### Create Issue
**Request:**
```bash
curl -X POST -H "PRIVATE-TOKEN: $GITCODE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Issue Title",
    "body": "Issue description",
    "labels": ["bug"]
  }' "${BASE_URL}/repos/{owner}/issues"
```
**Response:** 201
```json
{
  "id": 123,
  "iid": 1,
  "title": "Issue Title",
  "state": "opened",
  "number": 1,
  "created_at": "2026-04-02T00:00:00Z"
}
```

### Create MR
**Request:**
```bash
curl -X POST -H "PRIVATE-TOKEN: $GITCODE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "MR Title",
    "body": "## Summary\n- change",
    "head": "yuchaow:feature-branch",
    "base": "master"
  }' "${BASE_URL}/repos/{owner}/{repo}/pulls"
```
**Response:** 201
```json
{
  "id": 123,
  "iid": 1,
  "title": "MR Title",
  "state": "opened",
  "web_url": "https://gitcode.com/owner/repo/merge_requests/1"
}
```

### Get Repository Info
**Request:**
```bash
curl -H "PRIVATE-TOKEN: $GITCODE_TOKEN" \
  "${BASE_URL}/repos/{owner}/{repo}"
```
**Response:** 200
```json
{
  "id": 123,
  "name": "repo",
  "full_name": "owner/repo",
  "private": false,
  "html_url": "https://gitcode.com/owner/repo",
  "default_branch": "master",
  "star_count": 10,
  "forks_count": 5
}
```

### List Commits
**Request:**
```bash
curl -H "PRIVATE-TOKEN: $GITCODE_TOKEN" \
  "${BASE_URL}/repos/{owner}/{repo}/commits?sha=master&page=1&per_page=20"
```
**Response:** 200
```json
{
  "data": [
    {
      "sha": "abc123",
      "short_sha": "abc123",
      "message": "commit message",
      "author": {
        "name": "Author",
        "email": "author@example.com",
        "date": "2026-04-02T00:00:00Z"
      },
      "url": "https://gitcode.com/owner/repo/commit/abc123"
    }
  ]
}
```

## Creating MRs

### Steps
1. Push branch to origin (your fork) first - **MR API requires branch exists**
2. Create MR via API to target upstream repo

```bash
# Push branch to your fork
git push -u origin feat/branch-name

# Create MR from fork to upstream
curl -X POST -H "PRIVATE-TOKEN: $GITCODE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "feat: description",
    "body": "## Summary\n- change",
    "head": "yuchaow:feat/branch-name",
    "base": "master"
  }' "https://api.gitcode.com/api/v5/repos/openeuler/yuanrong/pulls"
```

### Important Notes
- `head` format for fork MRs: `username:branch-name`
- Branch must exist on your origin BEFORE creating MR
- **WARNING**: Never push feature branches to upstream - only push to origin (your fork)
- For upstream repos: push to origin, then create MR to upstream

## Status Codes
- 200: Success (GET/PUT)
- 201: Created (POST)
- 204: No Content (DELETE)
- 401: Unauthorized
- 403: Forbidden
- 404: Not Found
- 429: Rate Limited
- 500: Server Error
