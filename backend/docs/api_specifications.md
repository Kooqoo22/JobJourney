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

---

## Profile

All profile endpoints require a valid Bearer access token.

---

### GET /api/v1/profile
Returns the authenticated user's profile.

**200 OK**
```json
{ "message": "profile retrieved", "data": { "id": 1, "email": "user@example.com", "full_name": "Jane Doe", "avatar_url": null, "timezone": "Asia/Jakarta", "auth_provider": "local", "is_verified": true, "role": "user", "created_at": "2026-07-09T10:00:00+07:00" } }
```

**401 Unauthorized** — missing/invalid token

**404 Not Found** — `{ "message": "user not found" }`

**500 Internal Server Error**

---

### PUT /api/v1/profile
Updates full name, avatar URL, and timezone. All three fields are required.

**Request**
```json
{ "full_name": "Jane Doe", "avatar_url": "https://cdn.example.com/avatar.jpg", "timezone": "Asia/Jakarta" }
```

**200 OK** — returns updated user object (same shape as GET /profile)

**400 Bad Request** — malformed JSON

**401 Unauthorized**

**404 Not Found** — `{ "message": "user not found" }`

**422 Unprocessable Entity** — `{ "message": "validation failed", "errors": [ { "field": "full_name", "message": "required" } ] }`

**500 Internal Server Error**

---

### PATCH /api/v1/profile/password
Changes the authenticated user's password. Requires current password verification.

**Request**
```json
{ "current_password": "oldpass1", "new_password": "newpass99" }
```

**200 OK**
```json
{ "message": "password changed" }
```

**400 Bad Request** — malformed JSON

**401 Unauthorized** — token invalid, or `{ "message": "current password is incorrect" }`

**422 Unprocessable Entity** — password policy: `{ "message": "password does not meet the requirements", "errors": [...] }`

**500 Internal Server Error**

---

### PATCH /api/v1/profile/preferences
Updates only the user's timezone preference.

**Request**
```json
{ "timezone": "America/New_York" }
```

**200 OK**
```json
{ "message": "preferences updated" }
```

**400 Bad Request** — malformed JSON

**401 Unauthorized**

**422 Unprocessable Entity** — `{ "message": "validation failed", "errors": [ { "field": "timezone", "message": "must be a valid IANA timezone" } ] }`

**500 Internal Server Error**

---

### DELETE /api/v1/profile
Cascade soft-deletes the authenticated user's account: application events, documents, job applications, refresh tokens, and the user record are all soft-deleted in a single transaction.

**200 OK**
```json
{ "message": "account deleted" }
```

**401 Unauthorized**

**404 Not Found** — `{ "message": "user not found" }`

**500 Internal Server Error**

---

## Job Applications

All application endpoints require a valid Bearer access token. Resources are scoped to the authenticated user — cross-user access returns `404` (anti-IDOR).

**Application object** (returned in `data`):
```json
{
  "id": 1,
  "company_name": "Acme Corp",
  "position_title": "Backend Engineer",
  "job_url": "https://jobs.acme.com/123",
  "work_arrangement": "remote",
  "employment_type": "full_time",
  "location": "Jakarta",
  "source": "LinkedIn",
  "status": "applied",
  "applied_date": "2026-07-01",
  "salary_min": "8000000.00",
  "salary_max": "12000000.00",
  "currency": "IDR",
  "notes": "Referred by John",
  "is_archived": false,
  "created_at": "2026-07-09T10:00:00+07:00",
  "updated_at": "2026-07-09T10:00:00+07:00"
}
```

Valid `status` values: `wishlist`, `applied`, `screening`, `interview`, `offer`, `accepted`, `rejected`, `withdrawn`, `ghosted`.

Valid `work_arrangement` values: `remote`, `onsite`, `hybrid`.

Valid `employment_type` values: `full_time`, `part_time`, `contract`, `internship`, `freelance`.

---

### POST /api/v1/applications
Creates a new job application. If `status` is not `wishlist` and `applied_date` is provided, the date must not be in the future. An `application_events` entry of type `applied` is auto-created when `status != wishlist`, inside the same transaction.

**Request**
```json
{
  "company_name": "Acme Corp",
  "position_title": "Backend Engineer",
  "job_url": "https://jobs.acme.com/123",
  "work_arrangement": "remote",
  "employment_type": "full_time",
  "location": "Jakarta",
  "source": "LinkedIn",
  "status": "applied",
  "applied_date": "2026-07-01",
  "salary_min": "8000000.00",
  "salary_max": "12000000.00",
  "currency": "IDR",
  "notes": "Referred by John"
}
```

Required: `company_name`, `position_title`. All other fields optional.

**201 Created** — returns the created application object.

**400 Bad Request** — malformed JSON

**401 Unauthorized**

**422 Unprocessable Entity** — `{ "message": "validation failed", "errors": [ { "field": "salary_min", "message": "must be less than or equal to salary_max" } ] }`

**422 Unprocessable Entity** — `{ "message": "applied_date cannot be in the future" }`

**500 Internal Server Error**

---

### GET /api/v1/applications
Lists the authenticated user's applications with cursor-based pagination and optional filters.

**Query Parameters**
| Parameter | Type | Description |
|---|---|---|
| `q` | string | Keyword search across company_name and position_title (case-insensitive) |
| `status` | string | Filter by status value |
| `source` | string | Filter by source |
| `work_arrangement` | string | Filter by work_arrangement |
| `employment_type` | string | Filter by employment_type |
| `from_date` | string | Applied date ≥ (YYYY-MM-DD) |
| `to_date` | string | Applied date ≤ (YYYY-MM-DD) |
| `is_archived` | bool | Filter archived (true) or active (false) |
| `sort_by` | string | `applied_date` (default) or `updated_at` or `created_at` |
| `sort_dir` | string | `desc` (default) or `asc` |
| `cursor` | string | Opaque pagination cursor |
| `limit` | int | Page size 1–100 (default 20) |

**200 OK**
```json
{ "message": "applications retrieved", "data": [ { ...application object... } ], "meta": { "next_cursor": "eyJ...", "has_next": true, "limit": 20 } }
```

**401 Unauthorized**

**500 Internal Server Error**

---

### GET /api/v1/applications/:id
Returns a single application by ID.

**200 OK** — returns the application object.

**401 Unauthorized**

**404 Not Found** — `{ "message": "application not found" }`

**500 Internal Server Error**

---

### PUT /api/v1/applications/:id
Updates an application. Uses optimistic concurrency: `updated_at` in the request must match the server-side value.

**Request** — all fields optional; omitted fields are left unchanged.
```json
{
  "company_name": "New Corp",
  "position_title": "Senior Engineer",
  "job_url": null,
  "work_arrangement": "hybrid",
  "employment_type": "full_time",
  "location": "Bandung",
  "source": "Referral",
  "applied_date": "2026-07-05",
  "salary_min": "10000000",
  "salary_max": "15000000",
  "currency": "IDR",
  "notes": "Updated notes",
  "updated_at": "2026-07-09T10:00:00+07:00"
}
```

**200 OK** — returns the updated application object.

**400 Bad Request** — malformed JSON

**401 Unauthorized**

**404 Not Found** — `{ "message": "application not found" }`

**409 Conflict** — `{ "message": "application was modified by another request, please refresh and try again" }` (stale `updated_at`)

**422 Unprocessable Entity** — validation errors (salary, date, etc.)

**500 Internal Server Error**

---

### DELETE /api/v1/applications/:id
Soft-deletes the application along with its events and documents in a single transaction.

**200 OK**
```json
{ "message": "application deleted" }
```

**401 Unauthorized**

**404 Not Found** — `{ "message": "application not found" }`

**500 Internal Server Error**

---

### PATCH /api/v1/applications/:id/restore
Restores a soft-deleted application if it was deleted within the last 30 days. Only the application record is restored (events and documents remain soft-deleted).

**200 OK** — returns the restored application object.

**401 Unauthorized**

**404 Not Found** — `{ "message": "application not found or not eligible for restore" }`

**409 Conflict** — `{ "message": "application can no longer be restored (deleted more than 30 days ago)" }`

**500 Internal Server Error**

---

### PATCH /api/v1/applications/:id/status
Changes the application status. No-op if the new status equals the current status. A `status_changed` event is auto-created inside the same transaction when the status actually changes.

**Request**
```json
{ "status": "interview" }
```

**200 OK** — returns the updated application object.

**400 Bad Request** — malformed JSON

**401 Unauthorized**

**404 Not Found** — `{ "message": "application not found" }`

**422 Unprocessable Entity** — invalid status value

**500 Internal Server Error**

---

### PATCH /api/v1/applications/:id/archive
Toggles the archived flag on an application.

**Request**
```json
{ "is_archived": true }
```

**200 OK**
```json
{ "message": "application archived" }
```

**400 Bad Request** — malformed JSON

**401 Unauthorized**

**404 Not Found** — `{ "message": "application not found" }`

**422 Unprocessable Entity** — `{ "message": "is_archived is required" }`

**500 Internal Server Error**

---

## Application Events

Events track timeline milestones for a job application. All endpoints require Bearer auth. Scoped to the authenticated user.

**Event object**:
```json
{
  "id": 1,
  "application_id": 1,
  "type": "interview",
  "title": "Technical Interview Round 1",
  "event_at": "2026-07-15T10:00:00+07:00",
  "notes": "Video call on Google Meet",
  "remind_at": "2026-07-15T09:00:00+07:00",
  "status_from": null,
  "status_to": null,
  "created_at": "2026-07-09T10:00:00+07:00",
  "updated_at": "2026-07-09T10:00:00+07:00"
}
```

Valid `type` values: `applied`, `phone_screen`, `interview`, `assessment`, `offer`, `follow_up`, `deadline`, `note`, `status_changed`.

---

### POST /api/v1/applications/:id/events
Creates an event for an application. The application must belong to the authenticated user.

**Request**
```json
{
  "type": "interview",
  "title": "Technical Interview Round 1",
  "event_at": "2026-07-15T10:00:00Z",
  "notes": "Video call on Google Meet",
  "remind_at": "2026-07-15T09:00:00Z"
}
```

Required: `type`, `title`, `event_at`.

**201 Created** — returns the created event object.

**400 Bad Request** — malformed JSON or invalid timestamp

**401 Unauthorized**

**404 Not Found** — `{ "message": "application not found" }` (application doesn't exist or belongs to another user)

**422 Unprocessable Entity** — validation errors

**500 Internal Server Error**

---

### GET /api/v1/applications/:id/events
Lists events for an application, ordered by `event_at` ascending then `id` ascending.

**Query Parameters**
| Parameter | Type | Description |
|---|---|---|
| `cursor` | string | Opaque pagination cursor |
| `limit` | int | Page size 1–100 (default 20) |

**200 OK**
```json
{ "message": "events retrieved", "data": [ { ...event object... } ], "meta": { "next_cursor": "eyJ...", "has_next": false, "limit": 20 } }
```

**401 Unauthorized**

**404 Not Found** — `{ "message": "application not found" }`

**500 Internal Server Error**

---

### PUT /api/v1/applications/:id/events/:event_id
Updates an event. All fields optional; omitted fields are left unchanged.

**Request**
```json
{
  "type": "interview",
  "title": "Updated Title",
  "event_at": "2026-07-16T10:00:00Z",
  "notes": "Rescheduled",
  "remind_at": null
}
```

**200 OK** — returns the updated event object.

**400 Bad Request** — malformed JSON

**401 Unauthorized**

**404 Not Found** — event not found or belongs to another user/application

**422 Unprocessable Entity** — validation errors

**500 Internal Server Error**

---

### DELETE /api/v1/applications/:id/events/:event_id
Soft-deletes an event.

**200 OK**
```json
{ "message": "event deleted" }
```

**401 Unauthorized**

**404 Not Found** — event not found or belongs to another user/application

**500 Internal Server Error**

---

## Statistics

All stats endpoints require a valid Bearer access token. Data is scoped to the authenticated user and excludes soft-deleted records.

---

### GET /api/v1/stats/summary
Dashboard summary: total count by status, upcoming events (next 7 days, up to 10), and 5 most recently updated applications. Excludes archived applications from counts.

**200 OK**
```json
{
  "message": "summary retrieved",
  "data": {
    "totals": {
      "all": 42,
      "by_status": { "applied": 10, "interview": 5, "offer": 2, "accepted": 1 }
    },
    "upcoming_events": [
      { "id": 3, "application_id": 1, "type": "interview", "title": "Technical Interview", "event_at": "2026-07-15T10:00:00+07:00" }
    ],
    "recent_applications": [
      { "id": 1, "company_name": "Acme Corp", "position_title": "Backend Engineer", "status": "interview", "updated_at": "2026-07-09T12:00:00+07:00" }
    ]
  }
}
```

**401 Unauthorized**

**500 Internal Server Error**

---

### GET /api/v1/stats/applications
Application analytics for a given period.

**Query Parameters**
| Parameter | Values | Description |
|---|---|---|
| `period` | `week`, `month` (default), `quarter`, `year` | Lookback window and trend granularity |

**Rate Definitions**
- `response_rate` — (screening + interview + offer + accepted + rejected + withdrawn) / total non-wishlist
- `interview_rate` — (interview + offer + accepted) / total non-wishlist
- `offer_rate` — (offer + accepted) / total non-wishlist

**200 OK**
```json
{
  "message": "statistics retrieved",
  "data": {
    "funnel": [ { "status": "applied", "count": 10 }, { "status": "interview", "count": 5 } ],
    "response_rate": 0.65,
    "interview_rate": 0.30,
    "offer_rate": 0.05,
    "trend": [ { "period": "2026-06", "count": 8 }, { "period": "2026-07", "count": 5 } ],
    "by_source": [ { "source": "LinkedIn", "count": 12 }, { "source": "unknown", "count": 3 } ]
  }
}
```

Trend period format: `YYYY-WNN` (week), `YYYY-MM` (month), `YYYY-QN` (quarter), `YYYY` (year).

**401 Unauthorized**

**422 Unprocessable Entity** — `{ "message": "period must be one of: week, month, quarter, year" }`

**500 Internal Server Error**

---

## Admin

All admin endpoints require a valid Bearer access token and `role = admin`. A non-admin token returns `403`.

---

### GET /api/v1/admin/users
Lists all user accounts with cursor-based pagination.

**Query Parameters**
| Parameter | Type | Description |
|---|---|---|
| `q` | string | Keyword search across email and full_name (case-insensitive) |
| `status` | string | `active` or `banned`; omit for all |
| `cursor` | string | Opaque pagination cursor |
| `limit` | int | Page size 1–100 (default 20) |

**Admin user object**:
```json
{
  "id": 1,
  "email": "user@example.com",
  "full_name": "Jane Doe",
  "avatar_url": null,
  "timezone": "Asia/Jakarta",
  "auth_provider": "local",
  "is_verified": true,
  "is_banned": false,
  "ban_reason": null,
  "banned_at": null,
  "last_login_at": "2026-07-09T10:00:00+07:00",
  "role": "user",
  "created_at": "2026-07-01T10:00:00+07:00"
}
```

**200 OK** — returns list of admin user objects with cursor meta.

**401 Unauthorized**

**403 Forbidden**

**500 Internal Server Error**

---

### DELETE /api/v1/admin/users/:id
Cascade soft-deletes a user account: application events, documents, job applications, refresh tokens, and the user record in a single transaction. Cannot delete yourself or another admin account.

**200 OK**
```json
{ "message": "user deleted" }
```

**401 Unauthorized**

**403 Forbidden** — insufficient role, or `{ "message": "cannot delete your own account" }`, or `{ "message": "cannot delete an admin account" }`

**404 Not Found** — `{ "message": "user not found" }`

**500 Internal Server Error**

---

### POST /api/v1/admin/users/:id/ban
Bans a user account and revokes all their active refresh tokens in a single transaction. Cannot ban yourself or another admin.

**Request** (body optional)
```json
{ "reason": "Violation of terms of service" }
```

**200 OK**
```json
{ "message": "user banned" }
```

**400 Bad Request** — malformed JSON

**401 Unauthorized**

**403 Forbidden** — insufficient role, or `{ "message": "cannot ban your own account" }`, or `{ "message": "cannot ban an admin account" }`

**404 Not Found** — `{ "message": "user not found" }`

**409 Conflict** — `{ "message": "user is already banned" }`

**500 Internal Server Error**

---

### DELETE /api/v1/admin/users/:id/ban
Lifts the ban on a user account. Cannot unban an admin account.

**200 OK**
```json
{ "message": "user unbanned" }
```

**401 Unauthorized**

**403 Forbidden** — insufficient role, or `{ "message": "cannot unban an admin account" }`

**404 Not Found** — `{ "message": "user not found" }`

**409 Conflict** — `{ "message": "user is not banned" }`

**500 Internal Server Error**
