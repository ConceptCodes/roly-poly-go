# Roly Poly

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Roly Poly is a Go REST API for creating polls, voting on poll options, and reporting poll results.

## Features

- API-key authentication with `x-api-key`
- User onboarding
- Poll create, list, read, update, close, and delete
- Public poll listing via `GET /api/polls?public=true`
- Atomic poll creation with options
- Single-option and multi-option vote support
- Vote counts and percentage reports
- Request validation and consistent JSON responses

## Prerequisites

[![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://golang.org/doc/install) version 1.21 or higher

PostgreSQL with the `gen_random_uuid()` function available.

## Configuration

The app loads configuration from `.env` and environment variables.

```env
PORT=8080
HTTP_TIMEOUT=15
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=postgres
DB_NAME=roly_poly
ENV=development
```

## Installation

```sh
git clone https://github.com/conceptcodes/roly-poly-go.git
cd roly-poly-go
go mod download
```

## Running

```sh
go run cmd/server/main.go
```

Verify that the service is running:

```sh
curl http://localhost:8080/api/health/alive
```

```json
{
  "message": "Service is alive",
  "data": null,
  "error_code": ""
}
```

## Authentication

Only health and onboarding endpoints are public. All poll endpoints require:

```http
x-api-key: <api_key>
```

Create a user and API key:

```sh
curl --location 'http://localhost:8080/api/onboard' \
  --header 'Content-Type: application/json' \
  --data '{
    "first_name": "Ada",
    "last_name": "Lovelace"
  }'
```

```json
{
  "message": "User onboarded successfully",
  "data": {
    "id": "user_uuid",
    "api_key": "api_key_uuid",
    "first_name": "Ada",
    "last_name": "Lovelace"
  },
  "error_code": ""
}
```

## Endpoints

| Method | Path | Auth | Description |
| --- | --- | --- | --- |
| GET | `/api/health/alive` | No | Liveness check |
| GET | `/api/health/status` | No | Readiness check |
| POST | `/api/onboard` | No | Create a user and API key |
| POST | `/api/polls` | Yes | Create a poll with options |
| GET | `/api/polls` | Yes | List polls owned by the authenticated user |
| GET | `/api/polls?public=true` | Yes | List public polls |
| GET | `/api/polls/{id}` | Yes | Fetch one poll |
| PATCH | `/api/polls/{id}` | Yes | Update a poll |
| DELETE | `/api/polls/{id}` | Yes | Delete a poll, options, and votes |
| PATCH | `/api/polls/{id}/close` | Yes | Close a poll |
| POST | `/api/polls/{id}/vote` | Yes | Cast one or more votes |
| GET | `/api/polls/{id}/report` | Yes | Get vote totals and percentages |

## Poll Examples

Create a poll:

```sh
curl --location 'http://localhost:8080/api/polls' \
  --header 'Content-Type: application/json' \
  --header 'x-api-key: <api_key>' \
  --data '{
    "title": "Lunch",
    "description": "Where should we eat?",
    "public": true,
    "options": ["Tacos", "Pizza", "Salad"]
  }'
```

List your polls:

```sh
curl --location 'http://localhost:8080/api/polls' \
  --header 'x-api-key: <api_key>'
```

List public polls:

```sh
curl --location 'http://localhost:8080/api/polls?public=true' \
  --header 'x-api-key: <api_key>'
```

Update a poll:

```sh
curl --location --request PATCH 'http://localhost:8080/api/polls/<poll_id>' \
  --header 'Content-Type: application/json' \
  --header 'x-api-key: <api_key>' \
  --data '{
    "title": "Team lunch",
    "description": "Where should we eat today?",
    "public": true
  }'
```

Close a poll:

```sh
curl --location --request PATCH 'http://localhost:8080/api/polls/<poll_id>/close' \
  --header 'x-api-key: <api_key>'
```

Delete a poll:

```sh
curl --location --request DELETE 'http://localhost:8080/api/polls/<poll_id>' \
  --header 'x-api-key: <api_key>'
```

## Voting And Reports

Cast a vote:

```sh
curl --location 'http://localhost:8080/api/polls/<poll_id>/vote' \
  --header 'Content-Type: application/json' \
  --header 'x-api-key: <api_key>' \
  --data '{
    "option_ids": ["option_uuid"]
  }'
```

API-created polls default to single-option voting. The model supports multi-option voting when `allow_multiple_votes` is true; for those polls, pass more than one option ID:

```json
{
  "option_ids": ["first_option_uuid", "second_option_uuid"]
}
```

Get a poll report:

```sh
curl --location 'http://localhost:8080/api/polls/<poll_id>/report' \
  --header 'x-api-key: <api_key>'
```

```json
{
  "message": "Poll report fetched successfully",
  "data": {
    "poll_id": "poll_uuid",
    "title": "Lunch",
    "total_votes": 3,
    "options": [
      {
        "option_id": "option_uuid",
        "label": "Tacos",
        "votes": 2,
        "percentage": 66.66666666666666
      }
    ]
  },
  "error_code": ""
}
```

## Response Shape

Successful and error responses use the same envelope:

```json
{
  "message": "Human readable message",
  "data": null,
  "error_code": ""
}
```

Known error codes:

| Code | HTTP status |
| --- | --- |
| `RP-400` | 400 Bad Request |
| `RP-401` | 401 Unauthorized |
| `RP-403` | 403 Forbidden |
| `RP-404` | 404 Not Found |
| `RP-500` | 500 Internal Server Error |

## Development

Run all packages:

```sh
go test ./...
```

## Roadmap

- [x] Add an endpoint to generate reports
- [x] Add more tests
