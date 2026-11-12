# Where shared code lives

Search by what the code does. Shared packages live directly under `internal/`.

| Package | What belongs here |
| --- | --- |
| `auth` | Staff and portal session tokens, recovery codes, verification token purposes |
| `permissions` | School permissions, role checks, portal course access |
| `schoollogs` | School activity log actions and writes |
| `schoollookup` | Resolving academic year and grade IDs to school IDs |
| `portalusers` | Account types and student/guardian school lookups |
| `uploads` | Upload statuses, categories, storage object writes |
| `mailer` | Sending email and queueing email |
| `cryptoutil` | SHA256 hashing and application encryption |
| `responsecodes` | Shared API error codes |
| `dbutil` | SQL result checks |
| `submissions` | Shared submission statuses |

Code used by one feature stays in that feature's package under
`application/`. A local `helpers.go` is fine there. When multiple features need
something, put it in the shared package that describes its purpose.

Keep staff and portal differences in clearly named files, like `auth/staff.go`
and `auth/portal.go`, rather than creating another catch-all helpers package.
