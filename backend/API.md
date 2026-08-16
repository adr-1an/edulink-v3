# EduLink API and setup guide

This document describes the Go API currently used by the EduLink staff frontend, along with the configuration required to run the application locally or deploy it.

> Last checked against the repository on July 22, 2026.

## Contents

- [Architecture](#architecture)
- [Requirements](#requirements)
- [Environment configuration](#environment-configuration)
- [Database setup](#database-setup)
- [Running the application](#running-the-application)
- [API conventions](#api-conventions)
- [Authentication](#authentication)
- [Permissions and role hierarchy](#permissions-and-role-hierarchy)
- [Shared response models](#shared-response-models)
- [Endpoint reference](#endpoint-reference)
- [Error codes](#error-codes)
- [Operational and deployment notes](#operational-and-deployment-notes)

## Architecture

The active application consists of:

- `backend/`: Go 1.25 HTTP API using Chi and PostgreSQL.
- `frontend/`: Next.js 16 application using React 19.
- `backend/db/migrations/`: ordered PostgreSQL migrations.
- `rs-backend/`: a separate, unfinished Rust implementation. It is not used by the current frontend and is not covered by this API reference.

The Go API exposes the staff application below `/v1/staff`. `/v1/student` and `/v1/guardian` are reserved but currently have no routes.

The default local URLs used throughout this guide are:

```text
Frontend: http://localhost:3000
API:      http://localhost:8080
```

## Requirements

- Go 1.25 or a compatible newer release
- PostgreSQL
- Node.js supported by Next.js 16
- npm
- SMTP credentials for registration, password reset, email-change, and staff-invitation emails
- Cloudflare Turnstile site and secret keys for the frontend authentication forms

## Environment configuration

### Backend

The backend always loads `backend/.env` at startup. Start with the following template:

```dotenv
# HTTP server
APP_PORT=8080
APP_HOST_OVERRIDE=
APP_NAME=EduLink
FRONTEND_URL=http://localhost:3000

# Sonyflake ID generation. Use a different integer for every concurrently
# running API instance.
MACHINE_ID=1

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_NAME=edulink
DB_USER=postgres
DB_PASS=postgres
SSL_MODE=disable

# Optional: replaces the DB_* connection assembled by the API.
DB_DSN_OVERRIDE=

# Used by the bundled dbox migration executable.
DB_TYPE=postgres
DB_DSN=postgres://postgres:postgres@localhost:5432/edulink?sslmode=disable
MIGRATION_DIR=db/migrations

# SMTP. The mail client currently connects to port 587.
SMTP_HOST=smtp.example.com
SMTP_USER=your-smtp-user
SMTP_PASS=your-smtp-password
SMTP_FROM=EduLink <no-reply@example.com>
```

Notes:

- `APP_HOST_OVERRIDE` replaces `:APP_PORT` completely when set. For example, `127.0.0.1:8080` binds only to localhost.
- `DB_DSN_OVERRIDE` is useful in production when a provider supplies one PostgreSQL connection URL.
- `SSL_MODE` defaults to `disable`. Production should normally use the mode required by the database provider, such as `require` or `verify-full`.
- `MACHINE_ID` is required. Sonyflake IDs can collide if multiple API instances use the same machine ID.
- `FRONTEND_URL` must be the externally reachable frontend origin because it is embedded in emailed links.
- SMTP delivery errors are generally logged by the API. Several endpoints still return success after scheduling an email, so logs and SMTP monitoring matter.

Do not commit `.env` files or real credentials.

### Frontend

Create `frontend/.env.local`:

```dotenv
# Server-side only. Do not add NEXT_PUBLIC_ to this value.
API_URL=http://localhost:8080

NEXT_PUBLIC_APP_NAME=EduLink

# Required by registration and login forms.
NEXT_PUBLIC_TURNSTILE_SITE_KEY=your-turnstile-site-key
TURNSTILE_SECRET_KEY=your-turnstile-secret-key
```

The frontend calls the API from server components and server actions. `API_URL` may therefore be an internal service URL in production, while `FRONTEND_URL` must remain public.

Turnstile is required by the current login and registration server actions. Missing keys cause those actions to fail rather than bypass verification.

## Database setup

### 1. Create the database

For a local PostgreSQL installation:

```bash
createdb edulink
```

Alternatively, create it with your PostgreSQL provider and set `DB_DSN`, `DB_DSN_OVERRIDE`, or the individual `DB_*` values accordingly.

### 2. Run migrations

The repository includes an ARM64 macOS `backend/dbox` executable. From `backend/`, initialize its migration table and apply every pending migration:

```bash
cd backend
chmod +x dbox
./dbox init
./dbox up
./dbox stat
```

Useful dbox commands:

```text
dbox make <name>       Create a migration
dbox up [-c N] [-v]   Apply pending migrations
dbox down [-c N]      Roll back migrations
dbox stat             Show migration status
dbox cleanup [-v]     Remove records whose migration directories are gone
```

The bundled binary is platform-specific and its source is not in this repository. On another platform, use a compatible dbox build or apply each `backend/db/migrations/*/up.sql` file in directory-name order with a migration runner that records completed migrations. Do not repeatedly run the raw files as a substitute for migration tracking: not every data migration is safely repeatable.

The migrations create users, sessions, schools, staff invitations, roles and permissions, academic years, grades, courses, posts, audit logs, students, course assignments, and supporting indexes. They also insert the reserved user with ID `0`, used when a post or audit-log author is deleted.

## Running the application

### Backend

```bash
cd backend
go mod download
go run .
```

Confirm it is reachable:

```bash
curl http://localhost:8080/ping
```

### Frontend

```bash
cd frontend
npm ci
npm run dev
```

Then open `http://localhost:3000`.

Production checks:

```bash
cd frontend
npm run lint
npx tsc --noEmit
npm run build
```

## API conventions

### Base URL and content type

All documented routes are relative to:

```text
http://localhost:8080/v1/staff
```

JSON request bodies require:

```http
Content-Type: application/json
```

The router sets `Content-Type: application/json` globally, including for responses with no body.

### Authentication header

Unless an endpoint is marked public, provide the session token returned by login:

```http
Authorization: Bearer <session-token>
```

The raw token is returned only once. The API stores its SHA-256 hash. The frontend stores the raw token in an HTTP-only, `SameSite=Lax` cookie named `token` and forwards it from server-side requests.

### IDs

Database IDs are signed 64-bit Sonyflake integers. Treat them as decimal strings in JavaScript to avoid precision loss. Most response structs intentionally encode IDs as JSON strings.

Path parameters such as `{schoolID}` and `{courseID}` are decimal integers. The `activeAcademicYearId` update field is currently decoded by Go as a JSON number, while course assignment uses a JSON string in `studentId`.

### Dates and nullable times

Normal timestamps are RFC 3339 strings, for example:

```json
"2026-07-22T14:30:00Z"
```

Course post `showUntil` and `editedAt` use Go's `sql.NullTime` representation:

```json
{"Time":"2026-07-30T18:00:00Z","Valid":true}
```

An absent value is represented as:

```json
{"Time":"0001-01-01T00:00:00Z","Valid":false}
```

The same shape is required when creating or updating a post. This is a current API quirk; plain `null` is not the contract used by the frontend.

Student dates of birth use `YYYY-MM-DD`.

### Empty collections

Some handlers build slices lazily. An empty collection may therefore be returned as `null` instead of `[]`. Clients should normalize both to an empty array.

### Common statuses

| Status | Meaning |
| --- | --- |
| `200 OK` | Successful read, login, or role update |
| `201 Created` | Resource created or assignment accepted |
| `204 No Content` | Successful mutation with no response body |
| `400 Bad Request` | Malformed JSON, unknown JSON field, or malformed numeric path ID |
| `401 Unauthorized` | Missing, invalid, expired, or revoked session; sometimes incorrect credentials |
| `403 Forbidden` | Missing permission, inaccessible resource, hierarchy violation, or privacy restriction |
| `404 Not Found` | Invalid token or selected resource in endpoints that expose this distinction |
| `409 Conflict` | Duplicate resource or invalid state transition |
| `410 Gone` | Expired invitation link |
| `422 Unprocessable Entity` | Well-formed request that fails validation |
| `500 Internal Server Error` | Database, ID generator, transaction, or internal failure |

Most errors have an empty body. Endpoints that need a more specific client message may return:

```json
{"code":"ERROR_CODE"}
```

All JSON decoders reject unknown request fields.

### Request limits and middleware

- Request timeout: 15 seconds
- Request ID, client IP, request logging, panic recovery, and heartbeat middleware are enabled.
- No CORS middleware is configured. The current frontend avoids cross-origin browser calls by using Next.js server components/actions. Add an explicit restricted CORS policy if a browser client will call the API directly.
- No API rate limiter is currently configured. Add rate limiting at the reverse proxy or application layer before exposing authentication and email endpoints publicly.

## Authentication

Registration is email-link based:

1. `POST /auth/register` stores a one-hour registration token and emails the frontend link.
2. `GET /auth/register/{token}` validates the token and reveals its email.
3. `POST /auth/register/{token}` consumes the token and creates the account.
4. `POST /auth/login` creates a session and returns the raw bearer token.

Session duration is one day by default or three calendar months when `stayLoggedIn` is true.

Registration and password-reset requests intentionally hide whether an email is registered in most cases.

## Permissions and role hierarchy

School owners pass every permission check for their school. Staff members receive permissions through roles. Frontend permission checks are only for presentation; the API remains authoritative.

The shared access object returned by many list endpoints is:

```json
{
  "owner": false,
  "roles": [
    {
      "position": 2,
      "permissions": ["course.list", "course.post.list"]
    }
  ]
}
```

Higher role positions outrank lower positions. A non-owner may only create, update, delete, grant permissions to, or assign roles below their own highest role position. Position `0` is the lowest role.

Current permission strings:

```text
school.view
school.update
school.promote
school.invite.list
school.invite.cancel

staff_helpers.view
staff_helpers.create
staff_helpers.delete
staff_helpers.role.add
staff_helpers.role.remove
staff_helpers.role.list

academicYear.create
academicYear.list
academicYear.toggleActive
academicYear.delete

grade.create
grade.list
grade.update
grade.delete

role.create
role.list
role.update
role.delete
role.permission.update

course.create
course.list
course.update
course.delete
course.post.create
course.post.list
course.post.update
course.post.delete
course.student.assign
course.student.remove
course.student.list

student.create
student.list
student.update
student.delete

log.list
```

`owner` is an internal permission sentinel and cannot be granted to a role.

## Shared response models

### User profile

```json
{
  "id": "123",
  "name": "Taylor Morgan",
  "email": "taylor@example.com",
  "phone": "+48 123 456 789",
  "publicProfile": false,
  "staffInvitationsDisabled": false,
  "updatedAt": "2026-07-22T14:30:00Z",
  "createdAt": "2026-07-20T10:00:00Z"
}
```

### Course

```json
{
  "id": "123",
  "name": "Mathematics",
  "description": "Core mathematics course",
  "color": "6366F1"
}
```

Colors follow a six-character hexadecimal convention without a leading `#`. The current API validates the length but does not validate that every character is hexadecimal, so clients should enforce the full format.

### Course post

```json
{
  "id": "123",
  "authorName": "Taylor Morgan",
  "title": "Room changed",
  "body": "Today's class is in room 204.",
  "showUntil": {"Time":"2026-07-23T16:00:00Z","Valid":true},
  "accentColor": "6366F1",
  "editedAt": {"Time":"0001-01-01T00:00:00Z","Valid":false},
  "createdAt": "2026-07-22T14:30:00Z"
}
```

### Access-bearing list response

Most school-scoped list endpoints follow this pattern:

```json
{
  "access": {"owner":true,"roles":[]},
  "courses": []
}
```

The collection property changes by endpoint: `academicYears`, `grades`, `courses`, `posts`, `students`, `staff`, `roles`, `logs`, or `invitations`.

## Endpoint reference

### Health

| Method | Path | Auth | Success |
| --- | --- | --- | --- |
| `GET` | `/ping` | Public | Heartbeat response |

### Auth

#### `POST /auth/register`

Public. Sends a one-hour registration link.

```json
{"email":"person@example.com"}
```

Email is trimmed, lowercased, and must be 5–254 characters. Returns `204`. Existing users and active registration tokens normally receive the same response to limit account enumeration.

#### `GET /auth/register/{token}`

Public. Returns `200` with:

```json
{"email":"person@example.com"}
```

Returns `404` when the token is missing, invalid, or expired.

#### `POST /auth/register/{token}`

Public. Consumes a registration token and creates the account.

```json
{
  "name": "Taylor Morgan",
  "phone": "+48 123 456 789",
  "password": "at-least-8-characters"
}
```

Name is 1–128 characters. Phone is optional, otherwise 3–32 characters. Password is at least 8 characters. Returns `201`, `404` for an invalid/expired token, or `422` for invalid fields.

#### `POST /auth/login`

Public.

```json
{
  "email": "person@example.com",
  "password": "at-least-8-characters",
  "stayLoggedIn": false
}
```

Returns `200`:

```json
{"token":"raw-session-token"}
```

Returns `401` for an unknown account or incorrect password.

#### `GET /auth`

Authenticated session check. Returns `204` when valid. Invalid tokens return `401`, possibly with `{"code":"INVALID_TOKEN"}`. A missing header returns `401` with `NO_TOKEN`.

#### `DELETE /auth/logout`

Revokes the current session. Returns `204`, or `401 INVALID_TOKEN` if it was already invalid/revoked.

#### `POST /auth/reset`

Public. Sends a password-reset email.

```json
{"email":"person@example.com"}
```

Returns `204` whether or not a valid account exists. Invalid email length returns `422 INVALID_EMAIL`.

#### `PUT /auth/reset/{token}`

Public. Consumes a password-reset token created within the last 24 hours, revokes all user sessions, and changes the password.

```json
{"newPassword":"at-least-8-characters"}
```

Returns `204`, `404 INVALID_TOKEN`, or `422`.

### Profile

| Method | Path | Permission | Body | Success |
| --- | --- | --- | --- | --- |
| `GET` | `/profile` | Authenticated user | — | `200 {"user": UserProfile}` |
| `PATCH` | `/profile` | Authenticated user | Profile update | `204` |
| `POST` | `/profile/email` | Authenticated user | Email-change request | `204` |
| `PUT` | `/profile/email/{token}` | Public token | — | `204` |
| `PUT` | `/profile/password` | Authenticated user | Password change | `204` |

Profile update:

```json
{
  "name": "Taylor Morgan",
  "phone": "+48 123 456 789",
  "publicProfile": true,
  "staffInvitationsDisabled": false
}
```

Name is at most 64 characters. Phone is empty or 3–64 characters.

Email-change request:

```json
{
  "newEmail": "new@example.com",
  "password": "current-password"
}
```

The email must be 5–254 characters. An incorrect password returns `401 INCORRECT_PASSWORD`. The verification token lasts 24 hours. Applying the token returns `409` if the address became occupied and `404` if the token is invalid or expired.

Password change:

```json
{
  "password": "current-password",
  "newPassword": "at-least-8-characters"
}
```

A successful password change revokes every session, including the current one.

### Schools

| Method | Path | Permission | Success |
| --- | --- | --- | --- |
| `GET` | `/schools` | Authenticated user | `200 {"schools": SchoolSummary[]}` |
| `POST` | `/schools` | Authenticated user | `201` |
| `GET` | `/schools/{schoolID}` | `school.view` | `200 {"school": SchoolDashboard}` |
| `PATCH` | `/schools/{schoolID}` | `school.update` | `204` |
| `DELETE` | `/schools/{schoolID}` | Owner only | `204` |
| `DELETE` | `/schools/{schoolID}/leave` | Non-owner staff member | `204` |

School summary:

```json
{"id":"123","ownerId":"456","name":"Example School","regionCode":"PL"}
```

Create:

```json
{"name":"Example School","regionCode":"PL"}
```

Name is 1–64 characters. Region is optional; otherwise it must parse as a two-letter country region. Validation may return `INVALID_NAME` or `INVALID_REGION_CODE`.

Update:

```json
{
  "name": "Example School",
  "regionCode": "PL",
  "activeAcademicYearId": 123456789
}
```

`activeAcademicYearId` is optional. Supplying it additionally requires `academicYear.toggleActive`, and the year must belong to the school. To clear the active year instead, use `PUT /schools/{schoolID}/academic-years`.

The dashboard response contains school metadata and only grades belonging to the active academic year:

```json
{
  "school": {
    "id": "123",
    "name": "Example School",
    "regionCode": "PL",
    "grades": [
      {
        "id": "456",
        "academicYearId": "789",
        "level": 2,
        "name": "2nd grade",
        "createdAt": "2026-07-22T14:30:00Z"
      }
    ],
    "createdAt": "2026-07-20T10:00:00Z",
    "updatedAt": "2026-07-22T14:30:00Z"
  }
}
```

Deleting a school is a soft delete. Leaving is unavailable to the owner and permanently removes the caller's staff membership; rejoining requires another invitation.

### Academic years and promotion

| Method | Path | Permission | Success |
| --- | --- | --- | --- |
| `GET` | `/schools/{schoolID}/academic-years` | `academicYear.list` | `200 {access, academicYears}` |
| `POST` | `/schools/{schoolID}/academic-years` | `academicYear.create` | `201` |
| `PUT` | `/schools/{schoolID}/academic-years` | `academicYear.toggleActive` | `204` |
| `DELETE` | `/academic-years/{yearID}` | `academicYear.delete` | `204` |
| `POST` | `/schools/{schoolID}/promote` | `school.promote` | `204` |

Academic year model:

```json
{"id":"123","startYear":2026,"endYear":2027,"isActive":true}
```

Create:

```json
{"academicYear":{"from":2026,"to":2027}}
```

`from` must be 1900–9999 and `to` cannot be earlier than `from`. Duplicate ranges return `409 ACADEMIC_YEAR_CONFLICT`.

`PUT /schools/{schoolID}/academic-years` has no body and clears the active year. It returns `500` when no row was changed, so callers should avoid invoking it when no active year exists.

An active academic year cannot be deleted (`409`). Existing grades also prevent deletion because the grade foreign key uses `ON DELETE RESTRICT`.

Promotion body:

```json
{
  "newAcademicYear": {"from":2027,"to":2028},
  "options": {
    "activateAfterPromotion": true,
    "transferGrades": true,
    "promoteGradeLevels": true
  }
}
```

Promotion requires an existing active year or returns `403 NO_ACTIVE_YEAR`. It can create and activate the new year, copy grades, and increment copied grade levels. Courses, students, teachers, and guardians are not transferred yet.

### Grades

| Method | Path | Permission | Success |
| --- | --- | --- | --- |
| `GET` | `/schools/{schoolID}/grades` | `grade.list` | `200 {access, grades}` |
| `POST` | `/academic-years/{yearID}/grades` | `grade.create` | `201` |
| `PATCH` | `/grades/{gradeID}` | `grade.update` | `204` |
| `DELETE` | `/grades/{gradeID}` | `grade.delete` | `204` |

Grade model:

```json
{"id":"123","name":"2nd grade","level":2}
```

Create and update use the same body:

```json
{"name":"{level}nd grade","level":2}
```

The name must contain `{level}`, must be 7–32 characters before substitution, and the level must be 0–20. The first `{level}` is replaced with the decimal level before storage. Validation codes are `MISSING_LEVEL_VAR`, `INVALID_NAME`, and `LEVEL_OUT_OF_RANGE`. A duplicate name/level in the same academic year returns `409`.

Deleting a grade cascades to its courses.

### Courses

| Method | Path | Permission | Success |
| --- | --- | --- | --- |
| `GET` | `/grades/{gradeID}/courses` | `course.list` | `200 {access, courses}` |
| `POST` | `/grades/{gradeID}/courses` | `course.create` | `201` |
| `PATCH` | `/courses/{courseID}` | `course.update` | `204` |
| `DELETE` | `/courses/{courseID}` | `course.delete` | `204` |

Create and update body:

```json
{
  "name": "Mathematics",
  "description": "Core mathematics course",
  "color": "6366F1"
}
```

Name is 1–32 characters, description is at most 128 characters, and color must contain exactly six characters after trimming an optional leading `#`. Course names are case-insensitively unique within a grade. Conflicts return `409`.

Deleting a course cascades to its posts and student assignments.

### Course posts

| Method | Path | Permission | Success |
| --- | --- | --- | --- |
| `GET` | `/courses/{courseID}/posts` | `course.post.list` | `200 {access, posts}` |
| `POST` | `/courses/{courseID}/posts` | `course.post.create` | `201` |
| `PATCH` | `/course-posts/{postID}` | `course.post.update` | `204` |
| `DELETE` | `/course-posts/{postID}` | `course.post.delete` | `204` |

Create and update body:

```json
{
  "title": "Room changed",
  "body": "Today's class is in room 204.",
  "accentColor": "6366F1",
  "showUntil": {"Time":"2026-07-23T16:00:00Z","Valid":true}
}
```

Title is 1–32 characters, body is 1–2048 characters, and accent color must contain six characters after trimming an optional `#`. Use `{"Time":"0001-01-01T00:00:00Z","Valid":false}` for no expiry.

The list response includes the author's display name rather than author ID. Expired posts are not currently filtered by the API; clients decide how to present them.

### Students

| Method | Path | Permission | Success |
| --- | --- | --- | --- |
| `GET` | `/schools/{schoolID}/students` | `student.list` | `200 {access, students}` |
| `POST` | `/schools/{schoolID}/students` | `student.create` | `201` |
| `PATCH` | `/students/{studentID}` | `student.update` | `204` |
| `DELETE` | `/students/{studentID}` | `student.delete` | `204` |

Create and update body:

```json
{
  "name": "Jamie",
  "lastName": "Rivera",
  "dob": "2012-05-18",
  "email": "jamie@example.com",
  "phone": "+48 123 456 789",
  "notes": "Optional staff notes",
  "accountEnabled": true,
  "password": "at-least-8-characters"
}
```

Rules:

- First and last name: 3–32 characters each.
- Email: required, lowercased, 5–254 characters, and unique within the school.
- Date of birth: `YYYY-MM-DD` or `null`.
- Phone: empty or 3–32 characters.
- Notes: at most 2048 characters.
- Enabling an account during creation requires a password.
- A supplied password must be at least 8 characters.
- During update, an empty password preserves the existing password hash.

The current list response uses `last_name`, not `lastName`:

```json
{
  "id": "123",
  "name": "Jamie",
  "last_name": "Rivera",
  "dateOfBirth": "2012-05-18T00:00:00Z",
  "email": "jamie@example.com",
  "phone": "+48 123 456 789",
  "notes": "Optional staff notes",
  "accountEnabled": true,
  "createdAt": "2026-07-22T14:30:00Z",
  "archivedAt": null
}
```

Delete is permanent and cascades to course assignments.

### Course student assignments

| Method | Path | Permission | Success |
| --- | --- | --- | --- |
| `GET` | `/courses/{courseID}/students` | `course.student.list` | `200 {access, students}` |
| `POST` | `/courses/{courseID}/students` | `course.student.assign` | `201` |
| `DELETE` | `/courses/{courseID}/students` | `course.student.remove` | `204` |

The list returns every student in the course's school, not only assigned students:

```json
{
  "access": {"owner":true,"roles":[]},
  "students": [
    {
      "id": "123",
      "name": "Jamie",
      "lastName": "Rivera",
      "email": "jamie@example.com",
      "assigned": true
    }
  ]
}
```

Assign and remove use the same body:

```json
{"studentId":"123"}
```

The student must belong to the same school as the course. Both operations are idempotent at the database level: assigning an existing assignment or removing a missing assignment still returns success and does not create a duplicate audit log.

### Roles and role permissions

| Method | Path | Permission | Success |
| --- | --- | --- | --- |
| `GET` | `/roles/permissions` | Authenticated user | `200 {"permissions": string[]}` |
| `GET` | `/schools/{schoolID}/roles` | `role.list` | `200 {"roles": Role[]}` |
| `POST` | `/schools/{schoolID}/roles` | `role.create` + hierarchy | `201` |
| `PATCH` | `/roles/{roleID}` | `role.update` + hierarchy | `200` |
| `DELETE` | `/roles/{roleID}` | `role.delete` + hierarchy | `204` |
| `PUT` | `/roles/{roleID}/permissions` | `role.permission.update` + hierarchy | `204` |

Role model:

```json
{
  "id": "123",
  "position": 1,
  "name": "Teacher",
  "color": "6366F1",
  "createdAt": "2026-07-22T14:30:00Z",
  "permissions": ["course.list","course.post.list"]
}
```

Create/update body:

```json
{"name":"Teacher","position":1,"color":"6366F1"}
```

Names are 1–32 characters and unique per school. Colors are exactly six characters. Positions are contiguous. Creating at an occupied position or moving a role automatically shifts other roles. Create accepts positions from `0` through the current role count; update accepts an existing position.

Set or revoke a permission:

```json
{"permission":"course.post.create","allow":true}
```

Unknown permission strings return `422`.

### Staff and staff roles

| Method | Path | Permission | Success |
| --- | --- | --- | --- |
| `GET` | `/schools/{schoolID}/staff` | `staff_helpers.view` | `200 {access, staff}` |
| `DELETE` | `/staff-members/{staffID}` | Owner or `staff_helpers.delete` + hierarchy | `204` |
| `GET` | `/staff-members/{staffID}/roles` | `staff_helpers.role.list` | `200 {access, roles}` |
| `POST` | `/staff-members/{staffID}/roles/{roleID}` | `staff_helpers.role.add` + hierarchy | `204` |
| `DELETE` | `/staff-members/{staffID}/roles/{roleID}` | `staff_helpers.role.remove` + hierarchy | `204` |

Staff list entries contain the staff membership ID, user, inviter, creation time, and assigned roles:

```json
{
  "id": "123",
  "userId": "456",
  "user": {"id":"456","name":"Taylor Morgan","email":"taylor@example.com"},
  "addedBy": {"id":"789","name":"School Owner","email":"owner@example.com"},
  "createdAt": "2026-07-22T14:30:00Z",
  "roles": [{"id":"321","position":1,"name":"Teacher","color":"6366F1"}]
}
```

Non-owners cannot remove themselves through the staff-delete endpoint and can remove only staff whose highest role is below their own qualifying role. Staff can leave themselves through the school leave endpoint.

Role assignment/removal has no body. The staff member and role must belong to the same school.

### Staff invitations

| Method | Path | Auth/permission | Success |
| --- | --- | --- | --- |
| `GET` | `/staff-invitations` | Authenticated user | `200 {"invitations": Invitation[]}` |
| `GET` | `/schools/{schoolID}/staff-invitations` | `school.invite.list` | `200 {access, invitations}` |
| `POST` | `/schools/{schoolID}/staff/invitations` | `staff_helpers.create` | `204` |
| `GET` | `/staff-invitations/{token}` | Public token | `200 {"invitation": InvitationPreview}` |
| `POST` | `/staff-invitations/{token}/accept` | Intended authenticated user | `204` |
| `POST` | `/staff-invitations/{token}/reject` | Intended authenticated user | `204` |
| `POST` | `/staff-invitations/by-id/{invitationID}/accept` | Intended authenticated user | `204` |
| `POST` | `/staff-invitations/by-id/{invitationID}/reject` | Intended authenticated user | `204` |
| `POST` | `/staff-invitations/{invitationID}/cancel` | `school.invite.cancel` | `204` |

Send invitation:

```json
{"email":"person@example.com","importance":"normal"}
```

Valid importance values are `non-urgent`, `low`, `normal`, `high`, and `urgent`. Invitations expire after seven days. A second invitation to the same school/email within 24 hours returns `409 INVITATION_EMAIL_CONFLICT`; inviting an existing member returns `409 STAFF_MEMBER_EMAIL_CONFLICT`.

If the target already has an EduLink account and disabled staff invitations, the endpoint returns `403 TARGET_PRIVACY_RESTRICTED`.

The authenticated user's invitation list contains only unexpired pending invitations addressed to their current email:

```json
{
  "id": "123",
  "school": {"id":"456","name":"Example School","regionCode":"PL"},
  "sentBy": {"id":"789","name":"School Owner","email":"owner@example.com"},
  "status": "pending",
  "createdAt": "2026-07-22T14:30:00Z",
  "expiresAt": "2026-07-29T14:30:00Z"
}
```

The school invitation list calls the sender property `addedBy` and includes `userEmail`.

The public token preview includes `sentToEmail`, `schoolName`, `sentByName`, `sentByEmail`, `status`, `createdAt`, and `expiresAt`. An expired link returns `410`; an unknown link returns `404 INVALID_TOKEN`.

Accepting requires the logged-in user's email to match the invitation email. Acceptance creates a school staff membership but does not automatically assign a role.

### Audit logs

| Method | Path | Permission | Success |
| --- | --- | --- | --- |
| `GET` | `/schools/{schoolID}/logs` | `log.list` | `200 {access, logs}` |

Only logs from the previous 30 days are returned. There is currently no API-side pagination or explicit ordering.

```json
{
  "id": "123",
  "user": {"id":"456","name":"Taylor Morgan","email":"taylor@example.com"},
  "action": "course.student.assign",
  "type": "create",
  "title": "Course student assigned.",
  "message": "A student was assigned to the course Mathematics.",
  "details": "Taylor Morgan assigned Jamie to the course Mathematics.",
  "createdAt": "2026-07-22T14:30:00Z"
}
```

Current log types are `create`, `edit`, `delete`, and `other`. Treat `action`, `title`, `message`, and `details` as server-provided data. If localization is added later, prefer rendering localized strings from stable `action` values and structured details instead of translating arbitrary English text.

## Error codes

| Code | Typical status | Meaning |
| --- | --- | --- |
| `INVALID_TOKEN` | `401` or `404` | Session, registration, verification, reset, or invitation token is invalid |
| `INVALID_EMAIL` | `422` | Email failed validation |
| `NO_TOKEN` | `401` or `422` | Required token was omitted |
| `INCORRECT_PASSWORD` | `401` | Current password did not match |
| `INVALID_REGION_CODE` | `422` | School region is not a valid country region |
| `INVALID_NAME` | `422` | School or grade name failed validation |
| `INVALID_MAIL_IMPORTANCE` | `422` | Invitation email importance is unsupported |
| `INVITATION_EMAIL_CONFLICT` | `409` | A recent invitation already exists |
| `STAFF_MEMBER_EMAIL_CONFLICT` | `409` | The invited email is already school staff |
| `ACADEMIC_YEAR_CONFLICT` | `409` | The same academic-year range already exists |
| `MISSING_LEVEL_VAR` | `422` | Grade name template does not include `{level}` |
| `LEVEL_OUT_OF_RANGE` | `422` | Grade level is outside 0–20 |
| `NO_ACTIVE_YEAR` | `403` | School promotion was requested without an active year |
| `TARGET_PRIVACY_RESTRICTED` | `403` | Invitation target disabled staff invitations |

## Operational and deployment notes

### Reverse proxy and TLS

The Go server uses plain HTTP. Terminate HTTPS at a trusted reverse proxy or platform load balancer. Preserve the intended client IP header carefully; login records `X-Forwarded-For` when present.

### Email

SMTP port `587` is currently hardcoded. Confirm that the SMTP provider supports authenticated submission on that port. Registration, password-reset, invitation, password-change, and successful email-change flows depend on mail delivery.

### Security checklist

- Use HTTPS for both frontend and API.
- Use a restricted production PostgreSQL account and an SSL-enabled connection.
- Keep `API_URL`, database credentials, SMTP credentials, and the Turnstile secret server-side.
- Set a unique `MACHINE_ID` for each API replica.
- Add edge/application rate limits for login and email-producing endpoints.
- Restrict CORS if direct browser API access is introduced.
- Monitor API and SMTP logs; asynchronous mail failures do not always change the HTTP response.
- Back up PostgreSQL before migrations and destructive school/student operations.

### Current API caveats

- Student and guardian route groups are placeholders only.
- The frontend contains a resend-verification action for `POST /v1/staff/auth/verifications`, but the Go API does not currently mount that endpoint. Login also does not currently return the `403` state that exposes this flow.
- Audit-log pagination is frontend-only, and the API returns up to 30 days in one response.
- Course post nullable times use the verbose `sql.NullTime` JSON object.
- Empty lists may be `null`.
- Student list responses use `last_name`, while course roster responses use `lastName`.
- Course-post expiry is stored and returned but not filtered out by the list query.
- The API has no built-in CORS or rate-limiting middleware.
- The bundled migration executable is an ARM64 macOS binary and has no source code in this repository.
