# EduLink backend API and setup guide

This document covers the current Go backend: setup, runtime services, configuration, authentication, permissions, uploads, and every mounted HTTP endpoint.

> Verified against the backend source on September 3, 2026. API version: `3.13.6`.

## Architecture

The backend is a Go 1.25 service using Chi and PostgreSQL. It also uses:

- S3-compatible object storage through the MinIO Go SDK;
- SMTP for account and invitation email;
- a separate email queue worker for bulk student activation email;
- Sonyflake IDs with database-backed machine-ID leases;
- ordered SQL migrations under `internal/db/migrations/`.

The API has two independently authenticated route families:

```text
/v1/staff   Staff accounts and school administration
/v1/portal  Student and guardian accounts
```

Student workflows are implemented. Guardian authentication is recognized, but guardian-specific school workflows are not currently implemented. The public heartbeat is:

| Method | Path    | Auth   | Success            |
|--------|---------|--------|--------------------|
| `GET`  | `/ping` | Public | Heartbeat response |

## Requirements

- Go 1.25 or a compatible newer release
- PostgreSQL
- An S3-compatible object store and private bucket
- An SMTP server accepting authenticated submission on port 587
- The bundled `dbox` executable or another tracked migration runner

The bundled `dbox` binary is macOS ARM64. Another platform requires a compatible build or migration runner.

## Environment configuration

Both binaries call `godotenv.Load()`. Running from `backend/` loads `backend/.env`.

```dotenv
# Application
APP_PORT=8080
APP_HOST_OVERRIDE=
APP_NAME=EduLink
APP_ENCRYPTION_KEY=
FRONTEND_URL=http://localhost:3000
TRUST_X_FORWARDED_FOR=false

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_NAME=edulink
DB_USER=postgres
DB_PASS=postgres
SSL_MODE=disable
DB_DSN_OVERRIDE=

# dbox
DB_TYPE=postgres
DB_DSN=postgres://postgres:postgres@localhost:5432/edulink?sslmode=disable
MIGRATION_DIR=internal/db/migrations

# SMTP and queue worker
SMTP_HOST=smtp.example.com
SMTP_FROM=EduLink <no-reply@example.com>
SMTP_USER=your-smtp-user
SMTP_PASS=your-smtp-password
EMAIL_QUEUE_RATE=5

# S3-compatible object storage
S3_ENDPOINT=localhost:9000
S3_ACCESS_KEY_ID=your-access-key
S3_SECRET_ACCESS_KEY=your-secret-key
S3_BUCKET=edulink
S3_SSL=false
```

Important behavior:

- `APP_PORT` is required unless `APP_HOST_OVERRIDE` supplies the complete listener address.
- `APP_ENCRYPTION_KEY` must be hexadecimal and decode to exactly 32 bytes. Generate one with `openssl rand -hex 32`. It encrypts TOTP secrets and must remain stable.
- `FRONTEND_URL` is the backend's public application-origin setting used to construct emailed links.
- `TRUST_X_FORWARDED_FOR=true` actually configures Chi to trust `CF-Connecting-IP`. Enable it only behind a proxy that removes and replaces that header.
- `DB_DSN_OVERRIDE`, when set, replaces the DSN assembled from the individual database values.
- `SSL_MODE` defaults to `disable`.
- `EMAIL_QUEUE_RATE` is the worker polling interval in seconds and defaults to `5`.
- SMTP port 587 is hardcoded. `SMTP_PORT` is not read.
- `S3_ENDPOINT` is passed directly to `minio.New` and normally contains a host and optional port without a scheme.
- `S3_SSL` enables TLS only when exactly `true`.
- `S3_REGION` is not currently read.
- Sonyflake machine IDs require no environment variable. Processes lease IDs through PostgreSQL for 60 seconds and renew every 20 seconds.

Never commit `.env`, database credentials, storage credentials, SMTP credentials, or `APP_ENCRYPTION_KEY`.

## Database migrations

From `backend/`:

```bash
chmod +x dbox
./dbox init
./dbox up
./dbox stat
```

Useful commands:

```text
./dbox make <name>       Create a migration
./dbox up [-c N] [-v]   Apply pending migrations
./dbox down [-c N]      Roll back migrations
./dbox stat             Show migration status
./dbox cleanup [-v]     Clean records for removed migration directories
```

The runner reads `DB_TYPE`, `DB_DSN`, and `MIGRATION_DIR`. With another runner, apply `internal/db/migrations/*/up.sql` in directory-name order and track completed migrations. Do not repeatedly execute all files because data migrations may not be idempotent.

## Running the backend

Run the API from `backend/`:

```bash
go mod download
go run ./cmd/api
```

Run the email queue worker separately:

```bash
go run ./cmd/email-queue-worker
```

Verify the heartbeat:

```bash
curl http://localhost:8080/ping
```

The included Compose configuration builds and runs both binaries:

```bash
docker compose up --build
```

The API is bound to `127.0.0.1:${APP_PORT:-8080}` by Compose.

Build and test commands:

```bash
go build ./cmd/api
go build ./cmd/email-queue-worker
go test ./...
```

## API conventions

### Base URLs

```text
Staff:  http://localhost:8080/v1/staff
Portal: http://localhost:8080/v1/portal
```

All endpoint tables below use absolute API paths. JSON bodies should use `Content-Type: application/json`.

### Authorization

Authenticated routes expect:

```http
Authorization: Bearer <raw-session-token>
```

Only the SHA-256 token hash is stored. Staff and portal tokens use different tables and are not interchangeable.

### IDs and dates

Resource IDs are signed 64-bit Sonyflake integers. Most responses encode IDs as decimal strings to avoid numeric precision loss:

```json
{"id":"123456789012345678"}
```

Path IDs are decimal integers. `studentId` and `referencedPostId` are JSON strings, while `activeAcademicYearId` is a JSON number or `null`.

Timestamps are RFC 3339. Nullable timestamps, including post `showUntil` and `editedAt`, are an RFC 3339 string or `null`. Student dates of birth are submitted as `YYYY-MM-DD`.

### Collections and decoding

Some lazily initialized slices encode an empty result as `null`; consumers should accept both `null` and `[]`. Most handlers reject unknown JSON fields with `400`. Bulk import, grading, and challenge completion currently use non-strict decoders.

### Middleware

The router enables request IDs, client-IP extraction, logging, panic recovery, a 15-second timeout, heartbeat handling, and a global JSON content-type header. It has no CORS or rate-limiting middleware.

### Common statuses

| Status | Meaning                                                                                       |
|--------|-----------------------------------------------------------------------------------------------|
| `200`  | Successful read/login/challenge or data-returning mutation                                    |
| `201`  | Resource, submission, or grade created                                                        |
| `204`  | Successful mutation with no body                                                              |
| `400`  | Malformed JSON, unknown field, or malformed numeric path ID                                   |
| `401`  | Missing/invalid session, incorrect credentials, disabled portal account, or expired challenge |
| `403`  | Missing permission, hierarchy failure, ownership failure, or inaccessible resource            |
| `404`  | Missing token/resource where the handler exposes that distinction                             |
| `409`  | Duplicate data or invalid state transition                                                    |
| `410`  | Expired invitation link                                                                       |
| `422`  | Valid JSON that fails validation                                                              |
| `500`  | Database, storage, configuration, or other internal failure                                   |
| `501`  | Unsupported 2FA challenge purpose                                                             |

Most errors have no body. Selected errors return `{"code":"ERROR_CODE"}`.

## Sessions and account types

Staff registration links expire after one hour. Staff password-reset and email-change tokens are accepted within 24 hours. Staff sessions last one day when `stayLoggedIn` is false and three calendar months when true. Password reset and password change revoke all staff sessions.

When staff 2FA is enabled, login returns a 15-minute challenge instead of a session. The same challenge endpoint completes 2FA-protected school deletion.

Portal login requires `account_enabled = true` and `account_active = true`. Portal authentication uses a rolling seven-day activity window based on `last_used_at`. Successful authenticated requests refresh it. Portal password changes do not revoke other sessions.

Portal account types are `student` and `guardian`. Token check and profile accept either. Course, assignment, and submission routes require a student unless explicitly stated otherwise.

## Permissions and hierarchy

School owners pass all school permission checks. Other staff receive permissions through roles. An access-bearing response looks like:

```json
{
  "access": {
    "owner": false,
    "roles": [
      {"position":4,"permissions":["course.list","course.post.list"]}
    ]
  }
}
```

Higher numeric role positions outrank lower positions. Role creation/update/deletion, permission editing, role assignment/removal, and non-owner staff deletion also apply hierarchy checks. `owner` is reserved and cannot be assigned through a role.

| Area             | Assignable permissions                                                                                                       |
|------------------|------------------------------------------------------------------------------------------------------------------------------|
| School           | `school.view`, `school.update`, `school.promote`                                                                             |
| Invitations      | `school.invite.list`, `school.invite.cancel`                                                                                 |
| Staff            | `staff.view`, `staff.create`, `staff.delete`, `staff.role.add`, `staff.role.remove`, `staff.role.list`                       |
| Academic years   | `academicYear.create`, `academicYear.list`, `academicYear.toggleActive`, `academicYear.delete`                               |
| Grades           | `grade.list`, `grade.create`, `grade.update`, `grade.delete`                                                                 |
| Roles            | `role.create`, `role.list`, `role.update`, `role.delete`, `role.permission.update`                                           |
| Courses          | `course.create`, `course.list`, `course.update`, `course.delete`                                                             |
| Course posts     | `course.post.create`, `course.post.list`, `course.post.view`, `course.post.update`, `course.post.delete`                     |
| Post attachments | `post.attachment.create`, `post.attachment.delete`                                                                           |
| Logs             | `log.list`                                                                                                                   |
| Students         | `student.create`, `student.list`, `student.view`, `student.update`, `student.delete`                                         |
| Course students  | `course.student.assign`, `course.student.remove`, `course.student.list`                                                      |
| Assignments      | `course.assignment.create`, `course.assignment.list`, `course.assignment.update`, `course.assignment.delete`                 |
| Submissions      | `submission.list`, `submission.view`, `submission.return`, `submission.delete`, `submission.grade`, `submission.removeGrade` |

## Upload protocol

Uploads are direct-to-object-storage:

1. Initialize with the original name, exact byte count, and declared content type.
2. `PUT` the bytes to the returned presigned URL with matching `Content-Type` metadata.
3. Complete using the returned object ID and `{"completionToken":"..."}`.

Initialization returns:

```json
{
  "id": "storage-object-id",
  "completionToken": "one-time-token",
  "url": "https://storage.example.com/presigned-put-url"
}
```

Submission uploads also return `attachmentId`.

Staff objects complete at `POST /v1/staff/uploads/{objectID}`. Portal submission objects complete at `POST /v1/portal/submissions/attachments/{objectID}`.

Completion checks owner, pending status, completion token, exact size, and content-type metadata. After a failed verification, the backend attempts cleanup and marks the row failed. It does not inspect file signatures or malware-scan content.

Presigned PUT URLs last five minutes. Presigned GET URLs generally last 15 minutes; staff submission view and portal profile-picture URLs currently last five minutes.

## Staff API

### Authentication and 2FA

| Method   | Path                                  | Auth            | Body                                              | Success                                     |
|----------|---------------------------------------|-----------------|---------------------------------------------------|---------------------------------------------|
| `POST`   | `/v1/staff/auth/register`             | Public          | `{"email":"person@example.com"}`                  | `204`                                       |
| `GET`    | `/v1/staff/auth/register/{token}`     | Public token    | None                                              | `200 {email}`                               |
| `POST`   | `/v1/staff/auth/register/{token}`     | Public token    | `StaffRegistration`                               | `201`                                       |
| `POST`   | `/v1/staff/auth/login`                | Public          | `StaffLogin`                                      | `200 {token}` or `200 {twoFactorChallenge}` |
| `GET`    | `/v1/staff/auth`                      | Staff           | None                                              | `204`                                       |
| `DELETE` | `/v1/staff/auth/logout`               | Staff           | None                                              | `204`                                       |
| `POST`   | `/v1/staff/auth/reset`                | Public          | `{"email":"person@example.com"}`                  | `204`                                       |
| `PUT`    | `/v1/staff/auth/reset/{token}`        | Public token    | `{"newPassword":"new-password"}`                  | `204`                                       |
| `POST`   | `/v1/staff/auth/two-factor`           | Staff           | None                                              | `200 {url}`                                 |
| `PUT`    | `/v1/staff/auth/two-factor`           | Staff           | `{"code":"123456"}`                               | `200 {codes}`                               |
| `DELETE` | `/v1/staff/auth/two-factor`           | Staff           | `{"code":"123456","password":"current-password"}` | `204`                                       |
| `DELETE` | `/v1/staff/auth/two-factor/recovery`  | Recovery code   | `{"recoveryCode":"..."}`                          | `204`                                       |
| `POST`   | `/v1/staff/auth/two-factor/challenge` | Challenge token | `TwoFactorCompletion`                             | `200 {token}` or `204`                      |

`StaffRegistration`:

```json
{"name":"Taylor Morgan","phone":"+48 123 456 789","password":"at-least-8-characters"}
```

Name is 1–128 characters, optional phone is 3–32, and password is at least eight. Registration start intentionally does not reveal existing accounts or active registration tokens.

`StaffLogin`:

```json
{"email":"person@example.com","password":"at-least-8-characters","stayLoggedIn":false}
```

2FA login response:

```json
{
  "twoFactorChallenge": {
    "token": "challenge-token",
    "purpose": "login",
    "expiresAt": "2026-09-03T14:00:00Z"
  }
}
```

Challenge completion is `{"code":"123456","challengeToken":"challenge-token"}`.

2FA setup returns an `otpauth://` URL. Verifying the six-digit TOTP enables 2FA and returns eight recovery codes as an object keyed by numeric strings. Recovery disables 2FA and deletes all recovery codes without requiring a session. Normal disable requires the password and TOTP.

### Staff profile

| Method   | Path                                | Auth         | Body                                                           | Success          |
|----------|-------------------------------------|--------------|----------------------------------------------------------------|------------------|
| `GET`    | `/v1/staff/profile`                 | Staff        | None                                                           | `200 {user}`     |
| `PATCH`  | `/v1/staff/profile`                 | Staff        | `StaffProfileUpdate`                                           | `204`            |
| `POST`   | `/v1/staff/profile/email`           | Staff        | `{"newEmail":"new@example.com","password":"current-password"}` | `204`            |
| `PUT`    | `/v1/staff/profile/email/{token}`   | Public token | None                                                           | `204`            |
| `PUT`    | `/v1/staff/profile/password`        | Staff        | `{"password":"current-password","newPassword":"new-password"}` | `204`            |
| `POST`   | `/v1/staff/profile/profile-picture` | Staff        | `UploadMetadata`                                               | `200 UploadInit` |
| `DELETE` | `/v1/staff/profile/profile-picture` | Staff        | None                                                           | `204`            |

`StaffProfileUpdate`:

```json
{"name":"Taylor Morgan","phone":"+48 123 456 789","publicProfile":true,"staffInvitationsDisabled":false}
```

Name is at most 64 characters. Phone is empty or 3–64. Profile output includes `profilePicture`, `id`, `name`, `email`, `phone`, `twoFactorStatus`, both privacy flags, and timestamps. `profilePicture` is `null` or `{presignedUrl,fileName,contentType}`. Status is `disabled`, `pending`, or `enabled`.

Email conflict intentionally receives a fake `204`; incorrect password returns `401 INCORRECT_PASSWORD`. Completing an email change can return `409` if the address became occupied. Password change revokes all staff sessions.

`UploadMetadata` is `{"name":"avatar.webp","declaredSize":245760,"declaredContentType":"image/webp"}`. Staff pictures allow JPEG, PNG, WebP, and GIF up to 5 MiB. Initialization replaces the database reference to an existing picture. Deleting a missing picture returns `404`.

### Schools, academic years, grades, and courses

| Method | Path | Permission | Body | Success |
| --- | --- | --- | --- | --- |
| `GET` | `/v1/staff/schools` | Staff | None | `200 {schools}` |
| `POST` | `/v1/staff/schools` | Staff | `SchoolMutation` | `201` |
| `GET` | `/v1/staff/schools/{schoolID}` | `school.view` | None | `200 {school}` |
| `PATCH` | `/v1/staff/schools/{schoolID}` | `school.update` | `SchoolUpdate` | `204` |
| `DELETE` | `/v1/staff/schools/{schoolID}` | Owner | None | `204` or `200 {twoFactorChallenge}` |
| `DELETE` | `/v1/staff/schools/{schoolID}/leave` | Non-owner member | None | `204` |
| `GET` | `/v1/staff/schools/{schoolID}/academic-years` | `academicYear.list` | None | `200 {access, academicYears}` |
| `POST` | `/v1/staff/schools/{schoolID}/academic-years` | `academicYear.create` | `AcademicYearCreate` | `201` |
| `PUT` | `/v1/staff/schools/{schoolID}/academic-years` | `academicYear.toggleActive` | None | `204` |
| `DELETE` | `/v1/staff/academic-years/{yearID}` | `academicYear.delete` | None | `204` |
| `POST` | `/v1/staff/schools/{schoolID}/promote` | `school.promote` | `SchoolPromotion` | `204` |
| `GET` | `/v1/staff/schools/{schoolID}/grades` | `grade.list` | None | `200 {access, grades}` |
| `POST` | `/v1/staff/academic-years/{yearID}/grades` | `grade.create` | `GradeMutation` | `201` |
| `PATCH` | `/v1/staff/grades/{gradeID}` | `grade.update` | `GradeMutation` | `204` |
| `DELETE` | `/v1/staff/grades/{gradeID}` | `grade.delete` | None | `204` |
| `GET` | `/v1/staff/grades/{gradeID}/courses` | `course.list` | None | `200 {access, courses}` |
| `POST` | `/v1/staff/grades/{gradeID}/courses` | `course.create` | `CourseMutation` | `201` |
| `PATCH` | `/v1/staff/courses/{courseID}` | `course.update` | `CourseMutation` | `204` |
| `DELETE` | `/v1/staff/courses/{courseID}` | `course.delete` | None | `204` |

Payloads:

```json
{"name":"North High School","regionCode":"PL"}
```

School name is 1–64. Region is normalized to uppercase and must be a country region; empty is allowed. Update additionally accepts `"activeAcademicYearId":123` or `null`. Changing it also requires `academicYear.toggleActive` and the year must belong to the school. The dedicated `PUT .../academic-years` route clears the active year.

```json
{"academicYear":{"from":2026,"to":2027}}
```

Start must be 1900–9999 and end cannot precede it. A duplicate range returns `409 ACADEMIC_YEAR_CONFLICT`. Active years cannot be deleted.

```json
{
  "newAcademicYear":{"from":2027,"to":2028},
  "options":{"activateAfterPromotion":true,"transferGrades":true,"promoteGradeLevels":true}
}
```

Promotion requires an active year. It can copy grades, increment their levels, and activate the new year. It does not transfer courses, students, staff, or guardians.

```json
{"name":"Grade {level}","level":7}
```

Grade create/update requires `{level}` in the template; only the first occurrence is replaced. Template length is 7–32 and level is 0–20.

```json
{"name":"Mathematics","description":"Algebra and geometry","color":"6366F1"}
```

Course name is required and at most 32, description at most 128, and color exactly six characters after optional `#` removal. Hex characters are not validated. Names are case-insensitively unique within a grade.

School deletion is soft deletion. If owner 2FA is enabled, deletion returns a `schoolDeletion` challenge; complete it at the shared challenge route. Grade and course deletion follow database cascades and permanently remove dependent data.

### Course roster

| Method   | Path                                    | Permission              | Body                  | Success                  |
|----------|-----------------------------------------|-------------------------|-----------------------|--------------------------|
| `GET`    | `/v1/staff/courses/{courseID}/students` | `course.student.list`   | None                  | `200 {access, students}` |
| `POST`   | `/v1/staff/courses/{courseID}/students` | `course.student.assign` | `{"studentId":"123"}` | `201`                    |
| `DELETE` | `/v1/staff/courses/{courseID}/students` | `course.student.remove` | `{"studentId":"123"}` | `204`                    |

The list contains every same-school student with `id`, `name`, `lastName`, `email`, and `assigned`. The selected student must belong to the course's school. Repeated assignment/removal is idempotent at the database layer.

### Course posts and attachments

| Method   | Path                                                | Permission               | Body                 | Success               |
|----------|-----------------------------------------------------|--------------------------|----------------------|-----------------------|
| `GET`    | `/v1/staff/courses/{courseID}/posts`                | `course.post.list`       | None                 | `200 {access, posts}` |
| `POST`   | `/v1/staff/courses/{courseID}/posts`                | `course.post.create`     | `CoursePostMutation` | `201 {id}`            |
| `GET`    | `/v1/staff/course-posts/{postID}`                   | `course.post.view`       | None                 | `200 {access, posts}` |
| `PATCH`  | `/v1/staff/course-posts/{postID}`                   | `course.post.update`     | `CoursePostMutation` | `204`                 |
| `DELETE` | `/v1/staff/course-posts/{postID}`                   | `course.post.delete`     | None                 | `204`                 |
| `POST`   | `/v1/staff/course-posts/{postID}/upload`            | `post.attachment.create` | `UploadMetadata`     | `200 UploadInit`      |
| `DELETE` | `/v1/staff/course-posts/attachments/{attachmentID}` | `post.attachment.delete` | None                 | `204`                 |

```json
{
  "title":"Room changed",
  "body":"Today's class is in room 204.",
  "accentColor":"6366F1",
  "showUntil":"2026-09-10T18:00:00Z"
}
```

Title and body are required. Title is at most 32, body at most 2,048, and color exactly six characters. `showUntil` may be `null`. Update sets `editedAt`.

List posts include `attachments`, `id`, `authorName`, nullable `authorProfilePictureURL`, title/body/color, `showUntil`, `editedAt`, and `createdAt`. Single-post output uses a one-element `posts` array and an `author` object with `id`, `name`, and `email`. All nullable post timestamps use the same `null` or RFC 3339 representation.

Post attachment metadata uses `name`, `declaredSize`, and `declaredContentType`. Limit is 5 MiB. Allowed: JPEG, PNG, GIF, WebP, ZIP, PDF. Complete through `/v1/staff/uploads/{objectID}`.

### Assignments and staff submission management

| Method | Path | Permission | Body | Success |
| --- | --- | --- | --- | --- |
| `GET` | `/v1/staff/courses/{courseID}/assignments` | `course.assignment.list` | None | `200 {access, assignments}` |
| `POST` | `/v1/staff/courses/{courseID}/assignments` | `course.assignment.create` | `AssignmentMutation` | `201` |
| `PATCH` | `/v1/staff/assignments/{assignmentID}` | `course.assignment.update` | `AssignmentMutation` | `204` |
| `DELETE` | `/v1/staff/assignments/{assignmentID}` | `course.assignment.delete` | None | Empty `200` currently |
| `GET` | `/v1/staff/assignments/{assignmentID}/submissions` | `submission.list` | None | `200 {access, submissions}` |
| `GET` | `/v1/staff/submissions/{submissionID}` | `submission.view` | None | `200 {access, submission}` |
| `DELETE` | `/v1/staff/submissions/{submissionID}/return` | `submission.return` | None | `204` |
| `DELETE` | `/v1/staff/submissions/{submissionID}` | `submission.delete` | None | `204` |
| `POST` | `/v1/staff/submissions/{submissionID}/grade` | `submission.grade` | `SubmissionGrade` | `201` |
| `DELETE` | `/v1/staff/submissions/{submissionID}/grade` | `submission.removeGrade` | None | `204` |

`AssignmentMutation`:

```json
{
  "referencedPostId":"123",
  "title":"Chapter review",
  "description":"Complete the exercises.",
  "dueDate":"2026-09-10T18:00:00Z",
  "submissionsEnabled":true,
  "submissionsCloseAt":"2026-09-12T18:00:00Z"
}
```

Reference and dates may be `null`. Title is 3–64; description is at most 4,096. On creation, dates cannot be past and close time cannot precede due time. Update intentionally skips those date-order checks. Reference must be in the same school on create and same course on update.

Assignment output includes nullable `referencedPost`, all mutation fields, and `createdAt`. Assignment deletion cascades to submissions, attachments, and grades. The current successful delete handler emits implicit empty `200`, not `204`.

Submission list includes `submitted` and `returned` rows with submitter identity and nullable notes. View includes completed attachments and nullable grade. Staff attachment fields are `id`, `presignedUrl`, `originalFilename`, and `declaredContentType`.

Returning only accepts `submitted`, removes an existing score, and sets `returned`; otherwise `409`. Deletion only accepts `returned`, removes database storage records, then attempts object deletion.

Grading body is `{"score":87,"notes":"Good work."}`. Score is 0–100 and notes at most 2,048. Pending/returned submissions return `409`. Regrading upserts the grader, score, notes, and timestamp. Grade removal is idempotent.

### Students and bulk import

| Method | Path | Permission | Body | Success |
| --- | --- | --- | --- | --- |
| `GET` | `/v1/staff/schools/{schoolID}/students` | `student.list` | None | `200 {access, students}` |
| `POST` | `/v1/staff/schools/{schoolID}/students` | `student.create` | `StudentMutation` | `201` |
| `POST` | `/v1/staff/schools/{schoolID}/students/import` | `student.create` | `StudentImport` | `201`; empty list is `204` |
| `GET` | `/v1/staff/students/{studentID}` | `student.view` | None | `200 {access, student, assignmentSubmissions}` |
| `PATCH` | `/v1/staff/students/{studentID}` | `student.update` | `StudentMutation` | `204` |
| `DELETE` | `/v1/staff/students/{studentID}` | `student.delete` | None | `204` |
| `POST` | `/v1/staff/students/{studentID}/profile-picture` | `student.update` | `UploadMetadata` | `200 UploadInit` |
| `DELETE` | `/v1/staff/students/{studentID}/profile-picture` | `student.update` | None | `204` |

`StudentMutation`:

```json
{
  "name":"Jamie",
  "lastName":"Rivera",
  "dob":"2012-05-18",
  "email":"jamie@example.com",
  "phone":"+48 123 456 789",
  "notes":"Optional internal note",
  "accountEnabled":true,
  "password":"at-least-8-characters"
}
```

First/last name are 3–32, email 5–254, phone empty or 3–32, notes at most 2,048, and DOB must parse. Create requires a password when enabled. Update with empty password preserves it; a supplied password must be at least eight characters and deletes the student's portal sessions.

List entries use `lastName` and include nullable `profilePictureURL`, student data, account state, and creation time. View returns `profilePicture` as `null` or `{presignedUrl}` and submitted assignment history with course, grade, nullable score, time, and notes.

`StudentImport`:

```json
{
  "students":[
    {"name":"Jamie","lastName":"Rivera","dateOfBirth":"2012-05-18","email":"jamie@example.com","phone":null,"notes":null}
  ],
  "enableAccounts":true
}
```

Maximum 5,000 rows. Invalid rows return `422` and may include zero-based `invalidRow`. Oversize returns `422 PAYLOAD_TOO_LONG`. Conflicts return `409 {"conflictingEmails":[...]}` with every matching email.

Enabled imports generate 24-hour activation tokens and queue email in `email_queue`; the worker must run to deliver it. Login remains blocked until activation. Student pictures allow JPEG, PNG, WebP, GIF up to 5 MiB. Existing picture returns `409`; removal is idempotent. Complete through the staff upload endpoint.

Student deletion is permanent and cascades through related sessions and schoolwork.

### Staff members, roles, and invitations

| Method   | Path                                                      | Permission                              | Body                                               | Success                        |
|----------|-----------------------------------------------------------|-----------------------------------------|----------------------------------------------------|--------------------------------|
| `GET`    | `/v1/staff/schools/{schoolID}/staff`                      | `staff.view`                            | None                                               | `200 {access, staff}`          |
| `DELETE` | `/v1/staff/staff-members/{staffID}`                       | Owner or `staff.delete` plus hierarchy  | None                                               | `204`                          |
| `GET`    | `/v1/staff/staff-members/{staffID}/roles`                 | `staff.role.list`                       | None                                               | `200 {access, roles}`          |
| `POST`   | `/v1/staff/staff-members/{staffID}/roles/{roleID}`        | `staff.role.add` plus hierarchy         | None                                               | `204`                          |
| `DELETE` | `/v1/staff/staff-members/{staffID}/roles/{roleID}`        | `staff.role.remove` plus hierarchy      | None                                               | `204`                          |
| `GET`    | `/v1/staff/roles/permissions`                             | Staff                                   | None                                               | `200 {permissions}`            |
| `GET`    | `/v1/staff/schools/{schoolID}/roles`                      | `role.list`                             | None                                               | `200 {roles}`                  |
| `POST`   | `/v1/staff/schools/{schoolID}/roles`                      | `role.create` plus hierarchy            | `RoleMutation`                                     | `201`                          |
| `PATCH`  | `/v1/staff/roles/{roleID}`                                | `role.update` plus hierarchy            | `RoleMutation`                                     | `200`                          |
| `DELETE` | `/v1/staff/roles/{roleID}`                                | `role.delete` plus hierarchy            | None                                               | `204`                          |
| `PUT`    | `/v1/staff/roles/{roleID}/permissions`                    | `role.permission.update` plus hierarchy | `{"permission":"course.post.create","allow":true}` | `204`                          |
| `GET`    | `/v1/staff/staff-invitations`                             | Staff                                   | None                                               | `200 {invitations}`            |
| `GET`    | `/v1/staff/schools/{schoolID}/staff-invitations`          | `school.invite.list`                    | None                                               | `200 {access, invitations}`    |
| `POST`   | `/v1/staff/schools/{schoolID}/staff/invitations`          | `staff.create`                          | `StaffInvitationCreate`                            | `204`; owner self-add is `201` |
| `GET`    | `/v1/staff/staff-invitations/{token}`                     | Public token                            | None                                               | `200 {invitation}`             |
| `POST`   | `/v1/staff/staff-invitations/{token}/accept`              | Intended staff account                  | None                                               | `204`                          |
| `POST`   | `/v1/staff/staff-invitations/{token}/reject`              | Intended staff account                  | None                                               | `204`                          |
| `POST`   | `/v1/staff/staff-invitations/by-id/{invitationID}/accept` | Intended staff account                  | None                                               | `204`                          |
| `POST`   | `/v1/staff/staff-invitations/by-id/{invitationID}/reject` | Intended staff account                  | None                                               | `204`                          |
| `POST`   | `/v1/staff/staff-invitations/{invitationID}/cancel`       | `school.invite.cancel`                  | None                                               | `204`                          |

Staff list includes membership ID, user ID/summary, inviter summary, creation time, roles, and nullable staff profile-picture URLs. A non-owner cannot delete themselves and must outrank the target's highest role.

`RoleMutation` is `{"name":"Teacher","position":2,"color":"6366F1"}`. Name is required and at most 32; color length is exactly six. Creation position is 0 through role count; update is 0 through count minus one. Reordering shifts surrounding roles. Names are school-unique.

Permission mutation accepts only values returned by `/roles/permissions`; allow/revoke is idempotent.

`StaffInvitationCreate` is `{"email":"teacher@example.com","importance":"high"}`. Accepted importance values are `""` (normal), `"non-urgent"`, `"low"`, `"high"`, and `"urgent"`. Literal `"normal"` is currently rejected because the mail library represents normal importance as an empty string.

Existing staff returns `409 STAFF_MEMBER_EMAIL_CONFLICT`; an invitation within the prior 24 hours returns `409 INVITATION_EMAIL_CONFLICT`; disabled invitations return `403 TARGET_PRIVACY_RESTRICTED`. Inviting the owner's own email creates membership immediately. Normal invitations expire after seven days.

User invitation list returns pending, unexpired invitations addressed to the current email. School list returns pending rows, including expired ones. Public token view returns `410` when expired.

### Audit logs and upload completion

| Method | Path | Permission | Body | Success |
| --- | --- | --- | --- | --- |
| `GET` | `/v1/staff/schools/{schoolID}/logs` | `log.list` | None | `200 {access, logs}` |
| `POST` | `/v1/staff/uploads/{objectID}` | Upload owner | `{"completionToken":"..."}` | `204` |

Logs are restricted to the previous 30 days, with no pagination and no explicit ordering. Each log includes `id`, acting `user`, `action`, `type`, `title`, `message`, `details`, and `createdAt`. Types are `create`, `edit`, `delete`, and `other`. The backend replaces `{user}` with the acting name when storing a log.

The upload endpoint completes staff profile pictures, staff-managed student pictures, and post attachments.

## Portal API

### Authentication and profile

| Method | Path | Account | Body | Success |
| --- | --- | --- | --- | --- |
| `POST` | `/v1/portal/auth/login` | Public | `{"email":"student@example.com","password":"password"}` | `200 {token}` |
| `GET` | `/v1/portal/auth` | Student or guardian | None | `200 {user}` |
| `DELETE` | `/v1/portal/auth` | Student or guardian | None | `204` |
| `POST` | `/v1/portal/auth/activate/{token}` | Public token | `{"newPassword":"new-password"}` | `200 {token}` |
| `PUT` | `/v1/portal/auth/password` | Student or guardian | `{"currentPassword":"old","newPassword":"new-password"}` | `204` |
| `GET` | `/v1/portal/profile` | Student or guardian | None | `200 {profile}` |

Login email is lowercased and must be 5–254. Unknown, inactive, disabled, or incorrect accounts return `401`.

Token check returns `{"user":{"id":123,"accountType":"student"}}`; this ID is currently a JSON number. Activation tokens are single-use and expire after 24 hours. New passwords require at least eight characters.

Profile contains nullable `pfpUrl`, `name`, `lastName`, `email`, `phone`, `dateOfBirth`, and school `{owner:{name,email},name,region}`. Its picture URL lasts five minutes. The portal profile API is currently read-only.

### Courses, posts, and assignments

| Method | Path | Account | Body | Success |
| --- | --- | --- | --- | --- |
| `GET` | `/v1/portal/courses` | Student | None | `200 {courses}` |
| `GET` | `/v1/portal/courses/{courseID}` | Assigned student | None | `200 {course, posts, assignments}` |
| `GET` | `/v1/portal/posts/{postID}` | Portal account with course access | None | `200 {post}` |
| `GET` | `/v1/portal/assignments` | Student | None | `200 {assignments}` |
| `POST` | `/v1/portal/assignments/{assignmentID}/submissions` | Assigned student | `{"notes":"Optional notes"}` | `201 {submissionId, status}` |

Course list includes assigned courses in the active academic year with grade, name, description, and accent color. Dashboard includes course/grade data, visible posts with authors and completed attachments, assignments, and joined submission attachments.

Dashboard posts exclude expired `show_until` rows. Direct post view does not apply that filter but still requires course access. Direct output includes attachments, author `{id,name,email}`, title/body/color, nullable dates, and creation time.

The global assignment list returns active-year assignments from assigned courses, referenced post content, the student's latest relevant submission, completed attachments, and grade/grader data when present.

Submission creation requires submissions to be enabled and before `submissionsCloseAt`. Notes are at most 2,048. It creates or resumes a `pending` draft. An incompatible existing active submission returns `409`.

Submission statuses:

```text
pending    Draft; attachments may be added or removed
submitted  Finalized
returned   Returned by staff and awaiting replacement/resubmission
```

### Submission uploads and finalization

| Method | Path | Account | Body | Success |
| --- | --- | --- | --- | --- |
| `POST` | `/v1/portal/submissions/{submissionID}/attachments` | Draft owner | `PortalUploadMetadata` | `200 UploadInitWithAttachmentId` |
| `POST` | `/v1/portal/submissions/attachments/{objectID}` | Upload owner | `{"completionToken":"..."}` | `204` |
| `DELETE` | `/v1/portal/submissions/{submissionID}/attachments/{attachmentID}` | Draft owner | None | `204` |
| `POST` | `/v1/portal/submissions/{submissionID}/submit` | Draft owner | None | `204` |

`PortalUploadMetadata`:

```json
{"fileName":"answer.pdf","declaredSize":1048576,"declaredContentType":"application/pdf"}
```

Maximum size is 50 MiB. Allowed types:

- JPEG, PNG, GIF, WebP;
- ZIP, gzip, and 7-Zip;
- PDF;
- DOC and DOCX;
- MP4 audio, WAV, MPEG audio;
- MPEG and WebM video.

Attachment changes require ownership, current course access, and `pending` status. Initialization also requires submissions to remain open. Missing attachment pair returns `404`, ownership/access mismatch `403`, and non-pending state `409`.

Finalization rejects closed submissions and incomplete uploads. Repeating finalization after the same row is submitted returns idempotent `204`. Successful resubmission deletes the older returned row for that student and assignment.

Portal assignment attachment output contains `id`, `fileName`, `fileSize`, `contentType`, and `presignedUrl`.

## Error codes

| Code                          | Typical status | Meaning                                      |
|-------------------------------|----------------|----------------------------------------------|
| `INVALID_TOKEN`               | `401`          | Session or workflow token is invalid         |
| `INVALID_RECOVERY_CODE`       | `404`          | Staff recovery code is invalid               |
| `EXPIRED_TOKEN`               | `401`          | Challenge or portal activation token expired |
| `INVALID_EMAIL`               | `422`          | Email validation failed                      |
| `NO_TOKEN`                    | `401`          | Required bearer/path token is missing        |
| `INCORRECT_PASSWORD`          | `401`          | Current password failed on email change      |
| `INVALID_REGION_CODE`         | `422`          | Region is not a country region               |
| `INVALID_NAME`                | `422`          | School or grade name failed validation       |
| `INVALID_MAIL_IMPORTANCE`     | `422`          | Invitation importance is unsupported         |
| `INVITATION_EMAIL_CONFLICT`   | `409`          | Recent invitation already exists             |
| `STAFF_MEMBER_EMAIL_CONFLICT` | `409`          | Email already belongs to school staff        |
| `ACADEMIC_YEAR_CONFLICT`      | `409`          | Academic-year range already exists           |
| `MISSING_LEVEL_VAR`           | `422`          | Grade template lacks `{level}`               |
| `LEVEL_OUT_OF_RANGE`          | `422`          | Grade level is outside 0–20                  |
| `NO_ACTIVE_YEAR`              | `403`          | Promotion requires an active year            |
| `TARGET_PRIVACY_RESTRICTED`   | `403`          | Target disabled staff invitations            |
| `PAYLOAD_TOO_LONG`            | `422`          | Bulk import exceeds 5,000 rows               |

Not every response with one of these statuses includes a code.

## Operational and security notes

- Terminate TLS at a trusted reverse proxy; the Go server itself uses HTTP.
- Do not trust forwarding headers from arbitrary clients. Staff login stores any non-empty `X-Forwarded-For`; portal login stores it only when it parses as one IP. Chi's optional trust setting uses `CF-Connecting-IP`.
- Keep the object bucket private and grant least-privilege credentials.
- Configure bucket CORS only as required for presigned operations, restricted by origin, method, and header.
- Add file-signature validation and malware scanning before considering uploads trusted.
- Add edge rate limits for login, registration, reset, activation, invitation, and 2FA routes.
- Keep bearer, activation, reset, invitation, upload completion, challenge, and recovery tokens out of logs.
- Monitor API, worker, and SMTP logs. Direct email failures are logged and frequently do not change the response.
- Back up PostgreSQL before migrations and destructive operations.
- School deletion is soft. Grade, course, assignment, student, and returned-submission deletion can cascade permanently.

The queue worker claims up to 100 due messages, retries with exponential backoff capped at 64 minutes, and marks a message failed after its maximum retries. The helper defaults `max_retries` to one when omitted.

## Current backend limitations

- Guardian-specific course, assignment, and relationship workflows are not implemented.
- There is no API pagination for lists.
- Logs cover 30 days but have no explicit ordering.
- Empty slices may encode as `null`.
- Staff post listing does not filter expired posts; portal dashboard listing does.
- Assignment update skips creation-time date-order validation.
- SMTP port is fixed at 587 and `S3_REGION` is unused.
