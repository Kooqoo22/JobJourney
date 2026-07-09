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

## Authentication

Email/password authentication with JWT access tokens and opaque, rotating refresh tokens. Google OAuth is planned but not yet implemented (`auth_provider` supports `local` and `google`; only `local` is issued today).

**Token model**
- **Access token** — JWT (HS256), carries `user_id` + `role`, TTL `15m` (`JWT_ACCESS_TTL`). Sent as `Authorization: Bearer <token>` on protected routes.
- **Refresh token** — opaque random string returned to the client; only its SHA-256 hash is stored. TTL `168h` (`JWT_REFRESH_TTL`). Rotated on every refresh (single-use). Presenting an already-revoked refresh token is treated as reuse: **all** of that user's refresh tokens are revoked and `401` is returned.
- **Email/reset tokens** — opaque random strings emailed as links; only the SHA-256 hash is stored. Verify TTL `24h` (`AUTH_VERIFY_TOKEN_TTL`), reset TTL `1h` (`AUTH_RESET_TOKEN_TTL`), both single-use.

**Password policy** — minimum 8 characters and must contain both letters and numbers. Enforced on register and reset.

**Verification gating** — accounts are created unverified. Verification is required only for email-based features; unverified users may access core features. Verification and password-reset emails are sent asynchronously (fire-and-forget); a mail delivery failure does not fail the request.

**Anti-enumeration** — login returns a single generic message regardless of whether the email exists or the password was wrong. `resend-verification` and `forgot-password` always return a generic success message even when the email is unknown, unverified state does not match, or the account is a non-local provider.

**User object** (returned in `data.user` / `data` where noted):
```json
{
  "id": 1,
  "email": "ada@example.com",
  "full_name": "Ada Lovelace",
  "avatar_url": null,
  "timezone": "Asia/Jakarta",
  "auth_provider": "local",
  "is_verified": false,
  "role": "user",
  "created_at": "2026-07-09T10:00:00+07:00"
}
```

**Tokens object** (returned in `data.tokens` / `data` where noted):
```json
{
  "access_token": "eyJhbGciOi...",
  "refresh_token": "H2f0h...",
  "token_type": "Bearer",
  "expires_in": 900
}
```

---

### POST /api/v1/auth/register
Creates a local account (unverified) and emails a verification link. `timezone` is optional (IANA name); defaults to the configured default when omitted.

- Auth: none

**Request**
```json
{ "email": "ada@example.com", "password": "hunter2go", "full_name": "Ada Lovelace", "timezone": "Asia/Jakarta" }
```

**201 Created**
```json
{ "message": "registration successful, please check your email to verify your account", "data": { "id": 1, "email": "ada@example.com", "full_name": "Ada Lovelace", "avatar_url": null, "timezone": "Asia/Jakarta", "auth_provider": "local", "is_verified": false, "role": "user", "created_at": "2026-07-09T10:00:00+07:00" } }
```

**400 Bad Request** — malformed JSON
```json
{ "message": "invalid request body" }
```

**409 Conflict**
```json
{ "message": "email is already registered" }
```

**422 Unprocessable Entity** — binding validation (missing/invalid fields)
```json
{ "message": "validation failed", "errors": [ { "field": "Email", "message": "must be a valid email address" } ] }
```

**422 Unprocessable Entity** — password policy
```json
{ "message": "password does not meet the requirements", "errors": [ { "field": "password", "message": "must be at least 8 characters" }, { "field": "password", "message": "must contain both letters and numbers" } ] }
```

**422 Unprocessable Entity** — invalid timezone
```json
{ "message": "validation failed", "errors": [ { "field": "timezone", "message": "must be a valid IANA timezone" } ] }
```

**500 Internal Server Error**
```json
{ "message": "internal server error" }
```

---

### POST /api/v1/auth/verify-email
Consumes a verification token and marks the account verified. Idempotent for an already-verified account only while the token is unused; a used token returns `400`.

- Auth: none

**Request**
```json
{ "token": "H2f0h9...raw-token" }
```

**200 OK**
```json
{ "message": "email verified successfully", "data": { "id": 1, "email": "ada@example.com", "full_name": "Ada Lovelace", "avatar_url": null, "timezone": "Asia/Jakarta", "auth_provider": "local", "is_verified": true, "role": "user", "created_at": "2026-07-09T10:00:00+07:00" } }
```

**400 Bad Request**
```json
{ "message": "verification link is invalid" }
```
```json
{ "message": "verification link has already been used" }
```
```json
{ "message": "verification link has expired" }
```

**403 Forbidden**
```json
{ "message": "your account has been disabled" }
```

**404 Not Found**
```json
{ "message": "account not found" }
```

**500 Internal Server Error**
```json
{ "message": "internal server error" }
```

---

### POST /api/v1/auth/resend-verification
Reissues a verification link. Always returns `200` with a generic message (anti-enumeration); does nothing for unknown, already-verified, banned, or non-local accounts.

- Auth: none

**Request**
```json
{ "email": "ada@example.com" }
```

**200 OK**
```json
{ "message": "if the email is registered and unverified, a verification link has been sent" }
```

**400 Bad Request** — malformed JSON
```json
{ "message": "invalid request body" }
```

**422 Unprocessable Entity**
```json
{ "message": "validation failed", "errors": [ { "field": "Email", "message": "must be a valid email address" } ] }
```

**500 Internal Server Error**
```json
{ "message": "internal server error" }
```

---

### POST /api/v1/auth/login
Authenticates and issues a session (access + refresh). Uses a single generic message for unknown email or wrong password.

- Auth: none

**Request**
```json
{ "email": "ada@example.com", "password": "hunter2go" }
```

**200 OK**
```json
{ "message": "login successful", "data": { "user": { "id": 1, "email": "ada@example.com", "full_name": "Ada Lovelace", "avatar_url": null, "timezone": "Asia/Jakarta", "auth_provider": "local", "is_verified": false, "role": "user", "created_at": "2026-07-09T10:00:00+07:00" }, "tokens": { "access_token": "eyJhbGciOi...", "refresh_token": "H2f0h...", "token_type": "Bearer", "expires_in": 900 } } }
```

**401 Unauthorized**
```json
{ "message": "email or password is incorrect" }
```

**403 Forbidden**
```json
{ "message": "your account has been disabled" }
```

**422 Unprocessable Entity**
```json
{ "message": "validation failed", "errors": [ { "field": "Email", "message": "must be a valid email address" } ] }
```

**500 Internal Server Error**
```json
{ "message": "internal server error" }
```

---

### POST /api/v1/auth/refresh
Rotates the refresh token and issues a new access token. The presented refresh token is revoked and replaced. Reuse of a revoked token revokes the whole token family.

- Auth: none (the refresh token in the body is the credential)

**Request**
```json
{ "refresh_token": "H2f0h...raw-token" }
```

**200 OK**
```json
{ "message": "token refreshed", "data": { "access_token": "eyJhbGciOi...", "refresh_token": "N3wR4w...", "token_type": "Bearer", "expires_in": 900 } }
```

**401 Unauthorized**
```json
{ "message": "invalid refresh token" }
```
```json
{ "message": "refresh token has been revoked" }
```
```json
{ "message": "refresh token has expired" }
```

**403 Forbidden**
```json
{ "message": "your account has been disabled" }
```

**400 Bad Request** — malformed JSON
```json
{ "message": "invalid request body" }
```

**500 Internal Server Error**
```json
{ "message": "internal server error" }
```

---

### POST /api/v1/auth/logout
Revokes a single refresh token. Idempotent — succeeds even if the token is unknown or already revoked.

- Auth: none

**Request**
```json
{ "refresh_token": "H2f0h...raw-token" }
```

**200 OK**
```json
{ "message": "logged out" }
```

**400 Bad Request** — malformed JSON
```json
{ "message": "invalid request body" }
```

**500 Internal Server Error**
```json
{ "message": "internal server error" }
```

---

### POST /api/v1/auth/logout-all
Revokes every active refresh token for the authenticated user.

- Auth: Bearer access token required

**200 OK**
```json
{ "message": "logged out from all devices" }
```

**401 Unauthorized**
```json
{ "message": "missing authorization token" }
```
```json
{ "message": "invalid or expired token" }
```

**500 Internal Server Error**
```json
{ "message": "internal server error" }
```

---

### POST /api/v1/auth/forgot-password
Emails a password-reset link. Always returns `200` with a generic message (anti-enumeration); does nothing for unknown, banned, or non-local/passwordless accounts.

- Auth: none

**Request**
```json
{ "email": "ada@example.com" }
```

**200 OK**
```json
{ "message": "if the email is registered, a password reset link has been sent" }
```

**400 Bad Request** — malformed JSON
```json
{ "message": "invalid request body" }
```

**422 Unprocessable Entity**
```json
{ "message": "validation failed", "errors": [ { "field": "Email", "message": "must be a valid email address" } ] }
```

**500 Internal Server Error**
```json
{ "message": "internal server error" }
```

---

### POST /api/v1/auth/reset-password
Consumes a reset token, sets a new password, and revokes all existing refresh tokens (forces re-login everywhere).

- Auth: none

**Request**
```json
{ "token": "H2f0h...raw-token", "new_password": "newpass99" }
```

**200 OK**
```json
{ "message": "password reset successful, please log in with your new password" }
```

**400 Bad Request**
```json
{ "message": "reset link is invalid" }
```
```json
{ "message": "reset link has already been used" }
```
```json
{ "message": "reset link has expired" }
```

**403 Forbidden**
```json
{ "message": "your account has been disabled" }
```

**404 Not Found**
```json
{ "message": "account not found" }
```

**422 Unprocessable Entity** — password policy
```json
{ "message": "password does not meet the requirements", "errors": [ { "field": "new_password", "message": "must contain both letters and numbers" } ] }
```

**500 Internal Server Error**
```json
{ "message": "internal server error" }
```

---

> Domain endpoints (applications, events, documents, statistics, admin) are added here as they are implemented, together with their Bruno request files, per the docs-are-part-of-done rule.
