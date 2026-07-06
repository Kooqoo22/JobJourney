# JobJourney API Specification

Single source of truth for API behavior. Code carries no comments — all behavioral documentation lives here. Keep this in sync with the Bruno collection (`backend/bruno/`) on every endpoint change.

---

## Conventions

### Base URL
- Business endpoints: `/api/v1`
- Operational endpoints: `/` (e.g. `/health`)

### Response Envelope
Every response uses a single flat envelope. Lists map the array directly to `data` and pagination to `meta` (never `data.data`).

```json
{ "message": "string", "data": {}, "meta": {}, "errors": [] }
```

`data`, `meta`, and `errors` are omitted when empty.

### Standard Headers
| Header | Direction | Purpose |
|---|---|---|
| `Authorization: Bearer <token>` | request | Access token on protected routes |
| `X-Timezone` | request | IANA timezone for presentation conversion (fallback `Asia/Jakarta`) |
| `X-Request-ID` | request/response | Correlation id; generated if absent |

### Status Code Contract
| Status | When |
|---|---|
| `200 / 201` | Success |
| `400 Bad Request` | JSON binding/parse failure, wrong type (syntax) |
| `401 Unauthorized` | Missing/invalid/expired token |
| `403 Forbidden` | Authenticated but wrong role |
| `404 Not Found` | Resource absent (also used for cross-user access, anti-IDOR) |
| `409 Conflict` | Valid payload violates DB state |
| `422 Unprocessable Entity` | Validation / business-semantic failure; uses `errors`, no `data` |
| `500 Internal Server Error` | DB / unexpected failure |

### Error Body Shapes
```json
{ "message": "invalid or expired token" }
```
```json
{ "message": "validation failed", "errors": [ { "field": "email", "message": "must be a valid email address" } ] }
```

### Pagination
All list endpoints use cursor + limit (offset is prohibited). Cursor is an opaque base64 token. An invalid/expired cursor is treated as "start from the beginning", never a 500. Meta shape:

```json
{ "next_cursor": "string", "has_next": true, "limit": 20 }
```

Default limit `20`, maximum `100`. Changing search/filter resets the cursor.

---

## Middleware Chain
Applied outermost → innermost, globally:

1. **RequestID** — attaches `X-Request-ID` to context + response header.
2. **Recovery** — recovers panics → `500` envelope, never leaks stack.
3. **Logger** — structured JSON log (method, path, status, latency, client ip, request id).
4. **CORS** — origins/methods/headers from env; preflight `OPTIONS` → `204`.
5. **ErrorHandler** — the single place domain errors become HTTP responses (reads `c.Errors`).
6. **Auth** (scoped) — JWT verify on protected groups; injects `user_id` + `role`.
7. **RequireRole** (scoped) — RBAC guard on admin routes.

---

## Endpoints

### GET /health
Liveness/readiness probe. Pings the database.

- Auth: none

**200 OK**
```json
{ "message": "service healthy", "data": { "status": "ok" } }
```

**503 Service Unavailable**
```json
{ "message": "database unavailable" }
```

---

> Domain endpoints (auth, applications, events, documents, statistics, admin) are added here as they are implemented, together with their Bruno request files, per the docs-are-part-of-done rule.
