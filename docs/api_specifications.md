# JobJourney API Specification

Base URL: `/api/v1`

All timestamps are returned in RFC 3339 format, converted to the user's stored timezone.
All protected endpoints require `Authorization: Bearer <access_token>`.
All requests/responses use `Content-Type: application/json`.
Standard envelope: `{ "message": string, "data"?: any, "meta"?: any, "errors"?: any }`.

---

## Health

### GET /health
Liveness/readiness probe.

**Responses**
- `200` `{ "message": "service healthy", "data": { "status": "ok" } }`
- `503` `{ "message": "database unavailable" }`

---

## Auth

### POST /api/v1/auth/register
Register a new local (email+password) account. Sends a verification email asynchronously.

**Request body**
```json
{ "email": "user@example.com", "password": "Secure123", "full_name": "Jane Doe", "timezone": "Asia/Jakarta" }
```
- `email` required, valid email, max 255 chars
- `password` required, 8–72 chars, must contain letters and digits
- `full_name` required, max 100 chars
- `timezone` optional, IANA tz name; defaults to server's `APP_DEFAULT_TIMEZONE`

**Responses**
- `201` `{ "message": "registration successful, please check your email to verify your account", "data": { "id": 1, "email": "user@example.com", "full_name": "Jane Doe", "avatar_url": null, "timezone": "Asia/Jakarta", "auth_provider": "local", "is_verified": false, "role": "user", "created_at": "2026-07-09T10:00:00+07:00" } }`
- `400` `{ "message": "invalid request body" }`
- `409` `{ "message": "email is already registered" }`
- `422` `{ "message": "password does not meet the requirements", "errors": [ { "field": "password", "message": "must be at least 8 characters" } ] }`
- `500` `{ "message": "internal server error" }`

---

### POST /api/v1/auth/verify-email
Verify email with the token received in the verification email. Token is single-use and expires after 24 hours.

**Request body**
```json
{ "token": "<raw-token-from-email>" }
```

**Responses**
- `200` `{ "message": "email verified successfully", "data": { ...user } }`
- `400` `{ "message": "verification link is invalid" }` — token not found
- `400` `{ "message": "verification link has already been used" }`
- `400` `{ "message": "verification link has expired" }`
- `400` `{ "message": "invalid request body" }`
- `403` `{ "message": "your account has been disabled" }` — user is banned
- `500` `{ "message": "internal server error" }`

---

### POST /api/v1/auth/resend-verification
Resend the email verification link. Previous active tokens for that user are invalidated. Returns the same message regardless of whether the email is registered (anti-enumeration).

**Request body**
```json
{ "email": "user@example.com" }
```

**Responses**
- `200` `{ "message": "if the email is registered and unverified, a verification link has been sent" }`
- `400` `{ "message": "invalid request body" }`
- `500` `{ "message": "internal server error" }`

---

### POST /api/v1/auth/login
Authenticate with email + password. Issues a short-lived JWT access token and a long-lived rotating opaque refresh token. Unverified users may log in but email-dependent features are restricted. Banned users are denied.

**Request body**
```json
{ "email": "user@example.com", "password": "Secure123" }
```

**Responses**
- `200` `{ "message": "login successful", "data": { "user": { ...userResponse }, "tokens": { "access_token": "eyJ...", "refresh_token": "H2f0...", "token_type": "Bearer", "expires_in": 900 } } }`
- `400` `{ "message": "invalid request body" }`
- `401` `{ "message": "email or password is incorrect" }` — unknown email, wrong password, or Google-only account
- `403` `{ "message": "your account has been disabled" }` — banned account
- `500` `{ "message": "internal server error" }`

---

### POST /api/v1/auth/refresh
Exchange a valid refresh token for a new access token. Refresh token is rotated on every call (token reuse detection: reusing a revoked token revokes the entire user session).

**Request body**
```json
{ "refresh_token": "<opaque-refresh-token>" }
```

**Responses**
- `200` `{ "message": "token refreshed", "data": { "access_token": "eyJ...", "refresh_token": "new-token", "token_type": "Bearer", "expires_in": 900 } }`
- `400` `{ "message": "invalid request body" }`
- `401` `{ "message": "invalid refresh token" }`
- `401` `{ "message": "refresh token has been revoked" }`
- `401` `{ "message": "refresh token has expired" }`
- `403` `{ "message": "your account has been disabled" }`
- `500` `{ "message": "internal server error" }`

---

### POST /api/v1/auth/logout
Revoke the provided refresh token. Idempotent — succeeds even if the token is already revoked or not found.

**Request body**
```json
{ "refresh_token": "<opaque-refresh-token>" }
```

**Responses**
- `200` `{ "message": "logged out" }`
- `400` `{ "message": "invalid request body" }`
- `500` `{ "message": "internal server error" }`

---

### POST /api/v1/auth/logout-all
**Protected.** Revoke all active refresh tokens for the authenticated user (log out from all devices).

**Responses**
- `200` `{ "message": "logged out from all devices" }`
- `401` `{ "message": "missing authorization token" }` / `{ "message": "invalid or expired token" }`
- `500` `{ "message": "internal server error" }`

---

### POST /api/v1/auth/forgot-password
Request a password reset link. Always returns the same message (anti-enumeration). Only for local (email+password) accounts; Google-only accounts are silently ignored. Link expires in 1 hour and is single-use.

**Request body**
```json
{ "email": "user@example.com" }
```

**Responses**
- `200` `{ "message": "if the email is registered, a password reset link has been sent" }`
- `400` `{ "message": "invalid request body" }`
- `500` `{ "message": "internal server error" }`

---

### POST /api/v1/auth/reset-password
Reset password using a valid reset token. On success, all active refresh tokens are revoked (force re-login on all devices).

**Request body**
```json
{ "token": "<raw-token-from-email>", "new_password": "NewSecure456" }
```

**Responses**
- `200` `{ "message": "password reset successful, please log in with your new password" }`
- `400` `{ "message": "invalid request body" }`
- `400` `{ "message": "reset link is invalid" }` — token not found
- `400` `{ "message": "reset link has already been used" }`
- `400` `{ "message": "reset link has expired" }`
- `403` `{ "message": "your account has been disabled" }`
- `422` `{ "message": "password does not meet the requirements", "errors": [ { "field": "new_password", "message": "..." } ] }`
- `500` `{ "message": "internal server error" }`

---

## Profile

### GET /api/v1/profile
**Protected.** Retrieve the authenticated user's profile.

**Responses**
- `200` `{ "message": "profile retrieved", "data": { "id": 1, "email": "user@example.com", "full_name": "Jane Doe", "avatar_url": null, "timezone": "Asia/Jakarta", "auth_provider": "local", "is_verified": true, "role": "user", "created_at": "2026-07-09T10:00:00+07:00" } }`
- `401` `{ "message": "missing authorization token" }`
- `404` `{ "message": "user not found" }`
- `500` `{ "message": "internal server error" }`

---

### PUT /api/v1/profile
**Protected.** Update the authenticated user's profile. Only provided fields are updated.

**Request body**
```json
{ "full_name": "Jane Smith", "avatar_url": "https://cdn.example.com/avatar.jpg", "timezone": "America/New_York" }
```
- `full_name` optional, max 100 chars
- `avatar_url` optional, valid URL or null to clear
- `timezone` optional, IANA tz name

**Responses**
- `200` `{ "message": "profile updated", "data": { ...profileResponse } }`
- `400` `{ "message": "invalid request body" }`
- `401` `{ "message": "missing authorization token" }`
- `422` `{ "message": "validation failed", "errors": [ ... ] }`
- `500` `{ "message": "internal server error" }`

---

### PATCH /api/v1/profile/password
**Protected.** Change password for local accounts. Requires current password verification. Revokes all other refresh tokens on success.

**Request body**
```json
{ "current_password": "OldPass123", "new_password": "NewPass456" }
```

**Responses**
- `200` `{ "message": "password changed successfully" }`
- `400` `{ "message": "invalid request body" }`
- `401` `{ "message": "missing authorization token" }` / `{ "message": "current password is incorrect" }`
- `403` `{ "message": "password cannot be changed for accounts that use Google sign-in" }`
- `422` `{ "message": "password does not meet the requirements", "errors": [ ... ] }`
- `500` `{ "message": "internal server error" }`

---

### PATCH /api/v1/profile/preferences
**Protected.** Update timezone and notification preferences.

**Request body**
```json
{ "timezone": "Asia/Tokyo", "email_reminders_enabled": true }
```

**Responses**
- `200` `{ "message": "preferences updated", "data": { "timezone": "Asia/Tokyo", "email_reminders_enabled": true } }`
- `400` `{ "message": "invalid request body" }`
- `401` `{ "message": "missing authorization token" }`
- `422` `{ "message": "validation failed", "errors": [ { "field": "timezone", "message": "must be a valid IANA timezone" } ] }`
- `500` `{ "message": "internal server error" }`

---

### DELETE /api/v1/profile
**Protected.** Permanently soft-delete the authenticated user's account. Cascade soft-deletes all their job applications, events, and documents in a single transaction. Revokes all refresh tokens.

**Responses**
- `200` `{ "message": "account deleted" }`
- `401` `{ "message": "missing authorization token" }`
- `500` `{ "message": "internal server error" }`

---

## Job Applications

### POST /api/v1/applications
**Protected.** Create a new job application for the authenticated user.

**Request body**
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
- `company_name` required, max 200 chars
- `position_title` required, max 200 chars
- `job_url` optional, valid URL
- `work_arrangement` optional, enum: `remote|onsite|hybrid`
- `employment_type` optional, enum: `full_time|part_time|contract|internship|freelance`
- `status` optional, default `applied`, enum: `wishlist|applied|screening|interview|offer|accepted|rejected|withdrawn|ghosted`
- `applied_date` optional, date string `YYYY-MM-DD`; required when status is not `wishlist`; must not be in the future (except `wishlist`)
- `salary_min`/`salary_max` optional decimal strings; `salary_min` must be ≤ `salary_max`
- `notes` optional, max 5000 chars

**Business rules**
- First `application_events` entry of type `applied` is auto-created when `status != wishlist`.

**Responses**
- `201` `{ "message": "application created", "data": { ...applicationResponse } }`
- `400` `{ "message": "invalid request body" }`
- `401` `{ "message": "missing authorization token" }`
- `422` `{ "message": "validation failed", "errors": [ ... ] }`
- `500` `{ "message": "internal server error" }`

---

### GET /api/v1/applications
**Protected.** List the authenticated user's job applications with cursor pagination.

**Query parameters**
- `q` — keyword search (company_name, position_title)
- `status` — filter by status enum value
- `source` — filter by source string
- `work_arrangement` — filter by enum value
- `employment_type` — filter by enum value
- `from_date` / `to_date` — filter by applied_date range (`YYYY-MM-DD`)
- `is_archived` — `true` or `false` (default excludes archived)
- `sort_by` — `applied_date|updated_at|company_name|status` (default: `updated_at`)
- `sort_dir` — `asc|desc` (default: `desc`)
- `cursor` — opaque cursor from previous response
- `limit` — 1–100, default 20

**Responses**
- `200` `{ "message": "applications retrieved", "data": [ ...applicationResponse ], "meta": { "next_cursor": "...", "has_next": true, "limit": 20 } }`
- `401` `{ "message": "missing authorization token" }`
- `500` `{ "message": "internal server error" }`

---

### GET /api/v1/applications/:id
**Protected.** Get a single application's full details. Returns 404 for another user's application (IDOR protection).

**Responses**
- `200` `{ "message": "application retrieved", "data": { ...applicationResponse } }`
- `401` `{ "message": "missing authorization token" }`
- `404` `{ "message": "application not found" }`
- `500` `{ "message": "internal server error" }`

---

### PUT /api/v1/applications/:id
**Protected.** Update a job application. All validated fields may be updated. Concurrent edit protection via `updated_at` check.

**Request body** — same fields as create (all optional)

**Responses**
- `200` `{ "message": "application updated", "data": { ...applicationResponse } }`
- `400` `{ "message": "invalid request body" }`
- `401` `{ "message": "missing authorization token" }`
- `404` `{ "message": "application not found" }`
- `409` `{ "message": "application was modified by another request, please refresh and retry" }` — stale `updated_at`
- `422` `{ "message": "validation failed", "errors": [ ... ] }`
- `500` `{ "message": "internal server error" }`

---

### DELETE /api/v1/applications/:id
**Protected.** Soft-delete an application. Cascade soft-deletes its events and documents in a single transaction.

**Responses**
- `200` `{ "message": "application deleted" }`
- `401` `{ "message": "missing authorization token" }`
- `404` `{ "message": "application not found" }`
- `500` `{ "message": "internal server error" }`

---

### PATCH /api/v1/applications/:id/restore
**Protected.** Restore a soft-deleted application within the 30-day retention window.

**Responses**
- `200` `{ "message": "application restored", "data": { ...applicationResponse } }`
- `401` `{ "message": "missing authorization token" }`
- `404` `{ "message": "application not found" }`
- `409` `{ "message": "application cannot be restored after the retention period" }`
- `500` `{ "message": "internal server error" }`

---

### PATCH /api/v1/applications/:id/status
**Protected.** Change application status. No-op if status unchanged. Auto-creates an `application_events` entry of type `status_changed`.

**Request body**
```json
{ "status": "interview" }
```

**Responses**
- `200` `{ "message": "status updated", "data": { ...applicationResponse } }`
- `400` `{ "message": "invalid request body" }`
- `401` `{ "message": "missing authorization token" }`
- `404` `{ "message": "application not found" }`
- `422` `{ "message": "validation failed", "errors": [ { "field": "status", "message": "must be one of: ..." } ] }`
- `500` `{ "message": "internal server error" }`

---

### PATCH /api/v1/applications/:id/archive
**Protected.** Toggle the archived state of an application.

**Request body**
```json
{ "is_archived": true }
```

**Responses**
- `200` `{ "message": "application archived" }` / `{ "message": "application unarchived" }`
- `400` `{ "message": "invalid request body" }`
- `401` `{ "message": "missing authorization token" }`
- `404` `{ "message": "application not found" }`
- `500` `{ "message": "internal server error" }`

---

## Application Events

### POST /api/v1/applications/:id/events
**Protected.** Add an event to an application's timeline.

**Request body**
```json
{
  "type": "interview",
  "title": "Technical Interview Round 2",
  "event_at": "2026-07-15T10:00:00+07:00",
  "notes": "Prepare system design",
  "remind_at": "2026-07-14T10:00:00+07:00"
}
```
- `type` required, enum: `applied|phone_screen|interview|assessment|offer|follow_up|deadline|note|status_changed`
- `title` required, max 200 chars
- `event_at` required, RFC 3339 datetime
- `notes` optional, max 2000 chars
- `remind_at` optional, RFC 3339 datetime; must be in the future

**Responses**
- `201` `{ "message": "event created", "data": { ...eventResponse } }`
- `400` `{ "message": "invalid request body" }`
- `401` `{ "message": "missing authorization token" }`
- `404` `{ "message": "application not found" }`
- `422` `{ "message": "validation failed", "errors": [ ... ] }`
- `500` `{ "message": "internal server error" }`

---

### GET /api/v1/applications/:id/events
**Protected.** Get the timeline of events for an application, sorted chronologically.

**Query parameters**
- `cursor` — opaque cursor
- `limit` — 1–100, default 50

**Responses**
- `200` `{ "message": "events retrieved", "data": [ ...eventResponse ], "meta": { "next_cursor": "...", "has_next": false, "limit": 50 } }`
- `401` `{ "message": "missing authorization token" }`
- `404` `{ "message": "application not found" }`
- `500` `{ "message": "internal server error" }`

---

### PUT /api/v1/applications/:id/events/:event_id
**Protected.** Update an event's details.

**Request body** — same fields as create (all optional)

**Responses**
- `200` `{ "message": "event updated", "data": { ...eventResponse } }`
- `400` `{ "message": "invalid request body" }`
- `401` `{ "message": "missing authorization token" }`
- `404` `{ "message": "event not found" }`
- `422` `{ "message": "validation failed", "errors": [ ... ] }`
- `500` `{ "message": "internal server error" }`

---

### DELETE /api/v1/applications/:id/events/:event_id
**Protected.** Soft-delete an event.

**Responses**
- `200` `{ "message": "event deleted" }`
- `401` `{ "message": "missing authorization token" }`
- `404` `{ "message": "event not found" }`
- `500` `{ "message": "internal server error" }`

---

## Statistics

### GET /api/v1/stats/summary
**Protected.** Dashboard summary: application count per status, upcoming events (next 7 days), recently updated applications.

**Responses**
- `200` `{ "message": "summary retrieved", "data": { "totals": { "all": 42, "by_status": { "applied": 10, "interview": 5, ... } }, "upcoming_events": [ ...eventResponse ], "recent_applications": [ ...applicationResponse ] } }`
- `401` `{ "message": "missing authorization token" }`
- `500` `{ "message": "internal server error" }`

---

### GET /api/v1/stats/applications
**Protected.** Application statistics: funnel rates, response/interview/offer rates, trend per week/month, breakdown by source.

**Query parameters**
- `period` — `week|month|quarter|year` (default: `month`) — determines trend granularity

**Responses**
- `200` `{ "message": "statistics retrieved", "data": { "funnel": [ ... ], "response_rate": 0.65, "interview_rate": 0.30, "offer_rate": 0.05, "trend": [ ... ], "by_source": [ ... ] } }`
- `401` `{ "message": "missing authorization token" }`
- `500` `{ "message": "internal server error" }`

---

## Admin

### GET /api/v1/admin/users
**Protected (admin only).** List user accounts with cursor pagination.

**Query parameters**
- `q` — search by email or full_name
- `status` — `active|banned` filter
- `cursor`, `limit`

**Responses**
- `200` `{ "message": "users retrieved", "data": [ ...adminUserResponse ], "meta": { ... } }`
- `401` `{ "message": "missing authorization token" }`
- `403` `{ "message": "insufficient permissions" }`
- `500` `{ "message": "internal server error" }`

---

### PATCH /api/v1/admin/users/:id/ban
**Protected (admin only).** Ban a user account. Revokes all active refresh tokens. Admins cannot ban themselves or other admins.

**Request body**
```json
{ "reason": "Violation of terms of service" }
```

**Responses**
- `200` `{ "message": "user banned" }`
- `400` `{ "message": "invalid request body" }`
- `401` `{ "message": "missing authorization token" }`
- `403` `{ "message": "insufficient permissions" }` / `{ "message": "cannot ban an admin account" }` / `{ "message": "cannot ban your own account" }`
- `404` `{ "message": "user not found" }`
- `409` `{ "message": "user is already banned" }`
- `500` `{ "message": "internal server error" }`

---

### PATCH /api/v1/admin/users/:id/unban
**Protected (admin only).** Remove a ban from a user account.

**Responses**
- `200` `{ "message": "user unbanned" }`
- `401` `{ "message": "missing authorization token" }`
- `403` `{ "message": "insufficient permissions" }` / `{ "message": "cannot unban an admin account" }`
- `404` `{ "message": "user not found" }`
- `409` `{ "message": "user is not banned" }`
- `500` `{ "message": "internal server error" }`

---

### DELETE /api/v1/admin/users/:id
**Protected (admin only).** Soft-delete a user account. Cascade soft-deletes all their data in one transaction. Revokes all refresh tokens. Admins cannot delete themselves, other admins, or the last remaining admin.

**Responses**
- `200` `{ "message": "user deleted" }`
- `401` `{ "message": "missing authorization token" }`
- `403` `{ "message": "insufficient permissions" }` / `{ "message": "cannot delete an admin account" }` / `{ "message": "cannot delete your own account" }`
- `404` `{ "message": "user not found" }`
- `500` `{ "message": "internal server error" }`
