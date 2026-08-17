# EduLink

EduLink is a school management and learning platform designed to connect school administrators, staff members, teachers, students, and, eventually, guardians in one structured system. It combines the administrative tools a school needs to organize its people and academic structure with a student-facing portal for courses, announcements, assignments, file submissions, feedback, and grades.

The project is organized as two connected applications:

- The **staff application** is the administrative workspace. It manages schools, academic years, grades, courses, staff, roles, permissions, students, course posts, assignments, submissions, grading, and audit history.
- The **portal application** is the learner and family-facing workspace. The student experience is implemented and includes assigned courses, posts, assignments, submissions, attachments, returned work, and grades. Guardian accounts are represented in the account model and routing architecture, but the guardian interface is currently a placeholder.

Although both applications use the same PostgreSQL database and API deployment, they have separate account records, authentication handlers, session tables, cookies, route trees, and user interfaces. This keeps staff administration isolated from portal access while allowing both sides to work with the same school and course data.

## Contents

- [Product goals](#product-goals)
- [Public website and onboarding](#public-website-and-onboarding)
- [Staff application](#staff-application)
- [Two-factor authentication](#two-factor-authentication)
- [Student portal](#student-portal)
- [Guardian portal status](#guardian-portal-status)
- [Localization, dates, and appearance](#localization-dates-and-appearance)
- [Loading, error, and interaction behavior](#loading-error-and-interaction-behavior)
- [Permission model](#permission-model)
- [Authentication and security architecture](#authentication-and-security-architecture)
- [File-storage architecture](#file-storage-architecture)
- [Email delivery](#email-delivery)
- [Data model](#data-model)
- [Technical stack](#technical-stack)
- [Current boundaries](#current-boundaries)

## Product goals

EduLink is intended to make common school workflows understandable without requiring users to learn database-style concepts or navigate a collection of disconnected tools. Its interface emphasizes:

- clear, permission-aware actions;
- strong warnings around destructive operations;
- responsive layouts for desktop and mobile;
- immediate loading, empty, error, and success feedback;
- direct navigation between related school records;
- a consistent visual language across administration and learning workflows;
- English and Polish interface support;
- localized dates displayed in the user's browser time zone;
- secure separation between staff and portal accounts.

## Public website and onboarding

The public-facing landing page introduces EduLink to schools, staff, students, and families. It uses an animated, modern presentation with interface-inspired demonstrations rather than externally generated product imagery. The page provides separate entry points for staff and portal users so that a student does not accidentally end up on the staff login screen.

The public experience includes:

- separate staff registration, staff login, and student/guardian portal login actions;
- an overview of the school, course, communication, assignment, and permission features;
- responsive navigation with a collapsible mobile menu;
- animated sections and transitions built with Framer Motion and GSAP;
- a language selector for English and Polish;
- links to the Privacy Policy and Terms of Service;
- dedicated legal-document layouts with section navigation on larger screens.

### Staff registration and account recovery

Staff registration begins with an email address. EduLink sends a time-limited registration link, and the recipient completes the account from the linked registration page. Registration pages explain that continuing means agreeing to the Terms of Service and Privacy Policy.

The staff authentication area also supports:

- email and password login;
- Cloudflare Turnstile verification on exposed authentication forms;
- password-reset email requests;
- token-based password reset;
- email-change verification links;
- persistent login selection;
- two-factor login challenges;
- two-factor recovery with a recovery code;
- logout and server-side session revocation.

### Portal login and account routing

Portal users sign in through a separate portal login page using credentials issued by their school. After authentication, the API returns the account type. The frontend routes students to `/app/portal/student` and guardians to `/app/portal/guardian`.

Disabled portal accounts are rejected during login and token validation. A disabled student therefore cannot continue retrieving protected portal data with an existing session.

## Staff application

### Staff home and school selection

The staff home page is the entry point to every school the user owns or can access. It loads profile, school, and invitation data together so the primary workspace can appear without a chain of client-side requests.

The page provides:

- separate groups for owned schools and schools accessed as a staff member;
- school search by name, region code, or localized region name;
- region flags and localized country names;
- school creation with a name and region;
- a pending-invitation notification count;
- an invitation drawer with loading, retry, empty, and error states;
- invitation acceptance and rejection;
- sender previews showing the inviter's name and email;
- invitation creation and expiration timestamps;
- a leave-school action for non-owner staff members;
- a confirmation explaining that leaving is immediate and rejoining requires a new invitation.

School owners cannot use the staff leave action. Ownership and staff membership are treated as different forms of access by the API.

### School dashboard

The school dashboard is centered on the school's active academic year and grade structure. It presents grades as the main path into courses and communicates when required setup is missing.

If a school has no active academic year, the dashboard shows a warning explaining that an academic year must be created and activated in school settings before grades can be created.

Grade management includes:

- listing the grades in the active academic year;
- creating, editing, and deleting grades when permitted;
- a guided grade-name builder for users who should not need to understand template variables;
- selectable label parts that show how the final grade name will look;
- support for ordinary labels and ordinal forms such as `2nd grade`;
- a manual-entry tab for advanced users who want to write a template using `{level}`;
- a live name preview;
- search and empty states;
- stronger deletion confirmation because deleting a grade cascades into its courses and their related content.

The sidebar navigation remains available throughout nested grade, course, student, assignment, and submission pages. On mobile, it becomes a collapsible navigation panel that mirrors the desktop destinations without permanently occupying screen space.

### Academic years and school settings

School settings provide both ordinary configuration and high-impact academic operations.

Administrators can:

- update the school name and region;
- create academic years with start and end years;
- view historical and current academic years;
- activate a selected academic year;
- clear the active academic year;
- delete an academic year when the backend's integrity requirements allow it;
- promote the school into a new academic year;
- optionally copy the existing grade structure during promotion;
- optionally increase copied grade levels;
- optionally activate the newly created academic year immediately.

Promotion uses a review step that summarizes the new year, whether it will become active, and whether grades will be transferred or promoted before the operation runs.

School deletion is intentionally difficult. The interface explains the impact, requires an acknowledgement, requires an exact typed phrase containing the school name, and, when two-factor authentication is enabled, completes deletion through a short-lived TOTP challenge. This reflects the fact that deleting a school cascades through most of its academic, membership, post, assignment, submission, attachment, and audit data.

### Courses

Courses belong to grades and inherit the school's academic-year context through that relationship. Each course has a name, description, and color used throughout its cards and dashboard.

Staff members with the appropriate permissions can:

- search courses within a grade;
- create and edit course details;
- open a course from its grade page;
- delete a course with an acknowledgement and exact-name confirmation;
- see clear empty and no-search-result states.

The course dashboard is designed to accommodate multiple course features rather than acting as a single post feed. It currently contains Posts and Assignments tabs, a course overview, at-a-glance information, recent activity, and student-roster access.

### Course roster and student assignment

The course roster connects school-level student records to individual courses.

The roster interface provides:

- separate Assigned and Available student tabs;
- search-friendly, compact student entries;
- assignment of a student to a course;
- removal of a student from a course;
- immediate local feedback while an assignment or removal is in progress;
- permission-aware action visibility;
- backend enforcement that the student and course belong to compatible school data.

Portal students only see courses to which they have been explicitly assigned.

### Course posts

Posts are the course communication layer. A post contains a title, body, optional accent color, author, creation timestamp, optional edit timestamp, and optional visibility deadline.

Staff post functionality includes:

- post creation, listing, full-page viewing, editing, and deletion;
- permission-specific controls for each operation;
- an optional **Show until** switch that clearly enables or disables the date-time field;
- author avatars and profile previews;
- author name and email information inside the profile popover rather than exposed as a loose link;
- an Edited indicator whose hover state shows the actual edit timestamp;
- original creation timestamps shown in the user's local time zone;
- strong deletion confirmation;
- immediate list updates after mutations rather than requiring a browser reload.

Posts support attachments during creation and editing. Existing attachments can be removed independently when the user has the post-attachment deletion permission.

### Post attachments and previews

Post attachments use a direct-to-object-storage workflow. The frontend requests an upload initialization record, uploads the file to a presigned S3-compatible URL, and then calls the completion endpoint so the API can validate and finalize the object record.

The attachment interface supports:

- preserving and displaying the original filename and content type;
- image thumbnails for JPEG, PNG, GIF, and WebP files;
- in-app PDF previews, including Safari-compatible rendering behavior;
- generic file-type representations for other supported formats;
- opening files in a separate browser tab;
- downloading an individual attachment;
- downloading every attachment with progress and partial-failure feedback;
- a full attachment viewer for posts with multiple files;
- deletion of individual post attachments;
- presigned URLs so storage objects do not need to be publicly readable.

The same viewer is reused for staff posts, student posts, and assignment-submission files to keep file behavior consistent.

### Assignments

Assignments belong to courses and can optionally reference an existing course post. This allows a teacher to keep a detailed announcement or resource post separate while linking it directly from the assignment.

Staff assignment tools include:

- creation, listing, editing, and permanent deletion;
- title and optional description;
- optional due date;
- a submissions-enabled setting;
- an optional submissions-closing timestamp;
- an optional referenced-post toggle;
- post selection with a quick preview to distinguish posts that share a title;
- a direct link to the referenced post;
- permission-aware submission-list access;
- a destructive warning explaining that deleting an assignment also removes its submissions, scores, and attachment relationships.

### Submission listing

Each assignment has a staff-facing submission list. The page presents the students who submitted work and separates active submissions from returned submissions.

The listing experience includes:

- submission status and received timestamps;
- student identity information;
- access to the full submission view;
- a control for including returned submissions;
- returned work hidden from the ordinary list unless explicitly requested;
- empty and permission-denied states;
- server-authoritative permissions for listing and opening submissions.

### Submission review, returns, and deletion

The staff submission page combines the student's work, identity, files, status, and grading information.

Staff can view:

- the student who submitted the work;
- the submission timestamp;
- student notes;
- every completed attachment;
- image and PDF previews;
- individual and bulk downloads;
- the current submitted or returned state;
- any existing score and teacher feedback.

Authorized staff can return submitted work to the student. Returned submissions appear as work requiring resubmission in the student portal. A returned submission can also be permanently deleted through a dedicated destructive confirmation. The delete operation is intentionally limited to returned work because it also removes associated attachment records and stored objects.

### Grading

Authorized staff can grade a submitted assignment with:

- an integer score from 0 to 100 percent;
- optional teacher feedback up to the backend's supported limit;
- a visible grading timestamp;
- the identity of the staff member who graded it;
- editing of an existing grade;
- removal of an existing grade when separately permitted.

Scores use a continuous red-to-green visual accent. Low results lean red, middle results transition through warmer colors, and results near 100 percent lean green. The color is used as a restrained highlight, border, score color, and progress indicator rather than as the only way to communicate the value.

### Student records

Student records are managed from a dedicated school page. Students are portal users with the `student` account type, school ownership, profile data, and an optional enabled login.

The student list includes:

- summary counts;
- search by student data;
- frontend pagination to keep large school lists manageable;
- profile links;
- creation and editing actions;
- account-enabled status;
- permission-aware controls;
- loading skeletons, empty states, no-match states, and API error handling;
- permanent deletion with an explicit warning about course assignments, submissions, files, and login access.

The student profile page provides a more complete record view. Depending on the viewer's permissions, it exposes edit and delete actions, contact details, school information, login status, and the student's assignment-submission history.

Students do not currently upload their own profile pictures. Staff profile pictures are supported separately.

### Bulk student import

Bulk import is designed for both small manual batches and large prepared files.

The import dialog has two modes:

1. **Enter manually** presents an editable student table where rows can be added or removed.
2. **Import a file** accepts CSV, semicolon-delimited CSV, TSV, and TXT data and includes a downloadable CSV template.

The import experience includes:

- support for first name, last name, date of birth, email, phone, and internal notes;
- a dedicated note viewer/editor so long notes do not make the table unusable;
- per-row validation and visible error explanations;
- required-column checks;
- duplicate-email detection within the imported data;
- handling of all email conflicts returned by the API;
- a maximum file-size check and row-count limit;
- pagination inside the dialog for large imports;
- row and issue summaries before submission;
- the option to enable student login accounts for the imported group.

When account creation is enabled, the backend generates a unique password for each student, hashes it before storage, and queues an email containing the initial portal credentials. The popup explains this behavior before the import runs.

### Staff members and invitations

The staff-management page lists school staff with their profile pictures, roles, and relevant actions. It also provides access to pending school invitations.

Invitation and staff functionality includes:

- inviting an existing staff user by email;
- choosing the importance of the invitation email notification;
- explaining that the importance field affects the outgoing email;
- respecting the recipient's staff-invitation privacy preference;
- presenting privacy-restricted and other failures as toasts;
- listing pending invitations;
- showing who sent each invitation, including name and email;
- showing when an invitation was sent;
- cancelling a pending invitation when permitted;
- removing a staff member from a school;
- viewing and changing a staff member's assigned roles;
- profile-picture display with a five-minute frontend cache bounded by the signed URL's real expiry.

### Roles, permissions, and hierarchy

EduLink uses school-specific roles rather than a single global staff type. A role has a name, color, ordered position, permission set, creator, and timestamps.

Role management supports:

- creating, renaming, recoloring, reordering, and deleting roles;
- drag-based role ordering;
- assigning and removing roles from staff members;
- permission categories for school settings, academic years, grades, courses, posts, post attachments, assignments, submissions, students, staff, invitations, roles, course rosters, and audit logs;
- presets such as administrator, academic manager, staff manager, and teacher;
- per-permission switches with pending and rollback behavior;
- localized permission labels and categories;
- highlighting destructive permissions;
- a hierarchy help popup that explains how role order affects management authority.

Role order has security meaning. A staff member generally cannot create, move, modify, assign, remove, or delete a role at or above the position of that member's highest role. Likewise, a lower-ranked user cannot manage a staff member whose highest role outranks them. School owners bypass the school role hierarchy.

Frontend checks hide or disable unavailable actions for clarity, but the Go API performs the authoritative permission and hierarchy checks.

### Audit logs

School audit logs record important administrative changes. A log contains a type, action identifier, human-readable title, message, expandable details, actor, and timestamp.

The audit-log page includes:

- clearer event titles such as school, role, staff, grade, course, post, and assignment changes;
- creation, edit, deletion, and other event categories;
- type-specific colors and icons;
- search across visible log data;
- type filtering;
- a compact table layout;
- expandable details when a log carries additional context;
- actor profile popovers with name and email;
- timestamps displayed in the user's local time zone;
- frontend pagination in groups of 25 so thousands of rows are not rendered simultaneously.

The surrounding audit-log interface is localized, but stored log titles, messages, and details are intentionally displayed as written by the backend and are not translated.

### Staff profile and preferences

The staff profile is organized into focused tabs rather than one long settings form.

Profile functionality includes:

- viewing and editing name and phone information;
- changing email through a verification link;
- changing the account password;
- uploading, replacing, and removing a profile picture;
- displaying the picture on staff lists, posts, invitations, grading details, and profile popovers;
- separate loading states for unrelated profile operations;
- privacy preferences for public profile visibility and whether staff invitations are allowed;
- interface-language selection;
- light and dark theme support;
- two-factor authentication setup and removal.

Portal student profiles are deliberately read-only. The portal explains that a student must contact a school administrator to correct personal data.

## Two-factor authentication

Staff accounts can enable time-based one-time password authentication from the Security section of the profile.

The setup flow includes:

- generation of a TOTP secret;
- a QR code for authenticator apps;
- a manual setup value;
- verification with a six-digit authenticator code;
- recovery-code generation and a dedicated one-time display step;
- support for recovery codes regardless of their display formatting;
- clear warnings to save recovery codes before closing the setup dialog.

After 2FA is enabled, staff login becomes a two-step challenge. The initial password check creates a short-lived challenge rather than a full session. Completing the TOTP challenge consumes it and finishes login. The login screen also offers recovery-code authentication.

Two-factor authentication can be disabled with the current password. It can also be recovered using a valid recovery code, after which the user is expected to authenticate again. TOTP secrets are encrypted with the application's encryption key before database storage, while recovery codes and challenge tokens are stored as hashes.

The same challenge infrastructure protects school deletion for 2FA-enabled owners.

## Student portal

### Student home

The student home page lists only courses assigned to the authenticated student. Courses are grouped by grade and sorted for easy scanning.

The page includes:

- course and grade summary metrics;
- search by course name, description, or grade;
- course cards using the course's accent color;
- grouped grade sections;
- clear empty, loading, network-error, and retry states;
- responsive portal navigation for Courses, Assignments, and Profile.

### Student course dashboard

The student course dashboard mirrors the structure of the staff dashboard without exposing management controls. It contains Posts and Assignments tabs and leaves room for additional course modules in the future.

Students can:

- review the course name, description, grade, and overview;
- browse currently visible course posts;
- open a post on a dedicated page;
- view author identity and profile information;
- see original and edited timestamps;
- preview and download post attachments;
- browse assignments belonging to the course;
- open referenced posts directly from assignments.

Posts whose visibility deadline has passed are excluded by the backend.

### All assignments workspace

The portal also provides a consolidated assignment page across every course assigned to the student. Its default presentation is a compact list, with an optional grid view.

Assignment organization includes:

- search by assignment, description, course, or referenced post;
- course filtering when viewing all assignments;
- sorting by due date or assignment date in ascending or descending order;
- separate To submit, Submitted, Graded, and Returned views;
- hiding the Returned view when no returned work exists;
- a clear graded state with the score shown directly in the list;
- relative due labels such as due today, due tomorrow, due in a number of days, or overdue;
- ordinary due-date presentation for work that was already submitted, avoiding an unnecessary overdue warning;
- a dedicated red past-due section only for work that has not been submitted;
- closed past-due work hidden by default, with an option to reveal it;
- a separate orange returned-work section;
- submission-closing timestamps;
- links back to the relevant course and referenced post;
- empty states for every filter and status combination.

### Creating a submission

Submission creation is intentionally staged. Beginning an assignment creates a submission in a `pending` state. While it remains pending, the student can update notes, add attachments, remove draft attachments, and recover from an interrupted upload. Only the final submit action changes the status to `submitted`.

The submission dialog supports:

- optional written notes;
- multiple attachments;
- file validation before upload;
- up to three file uploads running concurrently;
- per-file preparation, upload, verification, completion, and failure states;
- individual and overall progress indicators;
- a final confirmation explaining that submitted work becomes locked;
- retry-friendly behavior when one or more files fail;
- prevention of duplicate submit actions;
- clear feedback during slow uploads so students do not reload the page unnecessarily.

Once work is submitted, the student can view it but cannot add or remove attachments. This prevents the submitted record from changing silently after the teacher begins reviewing it.

### Returned work and resubmission

When staff returns a submission, it no longer appears as completed work. The portal displays it in a clearly marked returned section and offers resubmission if submissions are still open.

The resubmission flow shows the previous returned files for reference while creating a new pending submission. The database permits one active, non-returned submission per student and assignment while retaining returned attempts as separate historical records.

If submissions have closed, the portal still communicates the returned status but does not offer an invalid resubmit action.

### Grades and teacher feedback

Graded submissions receive a prominent result treatment rather than a small status badge. Students can open submission details to see:

- the submitted notes and files;
- the percentage score;
- red-to-green score styling;
- teacher feedback;
- when the work was graded;
- the name and profile preview of the staff member who graded it.

The portal distinguishes ungraded submitted work from graded work and gives each its own assignment filter.

### Student profile

The student profile displays the authenticated student's school-managed information. It is read-only and explains that changes must be requested from a school administrator. This avoids creating a second profile-editing path that could conflict with official school records.

## Guardian portal status

The data model recognizes `guardian` as a portal account type, authentication can route a guardian to the guardian application, and a guardian route shell exists. The actual guardian dashboard, student relationships, course visibility, assignment access, and guardian-specific permissions are not implemented yet.

The landing page may describe parent and guardian access as part of the intended product direction, but the repository's functional portal is currently student-focused.

## Localization, dates, and appearance

EduLink includes an application-level localization system rather than isolated translated components.

Current language support:

- English as the default language;
- Polish as the second language;
- an HTTP-only, one-year locale cookie;
- translated navigation, authentication, staff, student-portal, permission, loading, empty, error, and feedback strings across the main application;
- Polish-aware plural forms;
- localized country names;
- localized permission names, categories, and presets.

Some remaining content is intentionally or currently English-only, including backend-generated audit-log content, parts of the legal documents, and a small number of newer staff course-management strings.

Dates are stored and transferred as UTC timestamps but rendered in the user's browser time zone where practical. The date component uses a deterministic UTC fallback during server rendering and switches to browser-local formatting after hydration, preventing server/client text mismatches. English formatting uses US conventions rather than British locale formats.

Appearance preferences include light and dark themes stored in a cookie. The overall interface uses responsive cards, dialogs, drawers, sheets, tabs, skeleton loaders, badges, toasts, and carefully styled destructive confirmations. A global accessible context menu provides common copy, paste, cut, select-all, link, and image actions where appropriate.

## Loading, error, and interaction behavior

The frontend is built around explicit states rather than blank pages or silent failures.

Common behavior includes:

- route-level `loading.tsx` skeletons for staff and portal pages;
- server-rendered initial data for fast first paint;
- parallel data requests where pages require independent resources;
- permission-denied states distinct from ordinary empty states;
- retry actions for recoverable network failures;
- localized toast feedback for mutations;
- disabled and loading controls during requests;
- duplicate-submission prevention;
- local state updates followed by router refreshes when authoritative server data may have changed;
- confirmation dialogs for destructive or irreversible actions;
- typed confirmation and acknowledgement steps for cascading deletions;
- mobile layouts that collapse navigation and stack dense form actions;
- keyboard focus states and accessible labels throughout reusable controls.

## Permission model

School owners have full authority over their own school. Other staff receive capabilities through one or more ordered roles.

The current permission catalog covers:

- viewing, updating, and promoting school data;
- listing and cancelling school invitations;
- viewing, inviting, removing, and assigning roles to staff;
- creating, listing, activating, and deleting academic years;
- creating, listing, updating, and deleting grades;
- creating, listing, updating, and deleting courses;
- creating, listing, viewing, updating, and deleting course posts;
- creating and deleting post attachments;
- listing audit logs;
- creating, listing, viewing, updating, and deleting students;
- assigning, listing, and removing course students;
- creating, listing, updating, and deleting assignments;
- listing, viewing, returning, deleting, grading, and removing grades from submissions;
- creating, listing, updating, deleting, and configuring roles.

Access data returned with school and course responses is used by the frontend to present the right controls. It is not treated as the security boundary. Every protected backend handler validates the session, resource relationship, school ownership or required permission, and, where applicable, role hierarchy.

School deletion is owner-only and is not delegated through an ordinary role permission.

## Authentication and security architecture

### Staff security

Staff passwords use Argon2id hashing. Session tokens are random values returned only to the client, while the database stores SHA-256 token hashes. Staff session records include creation metadata, last-use time, expiry, revocation time, and optional revocation notes.

Staff authentication also includes:

- time-limited registration tokens;
- time-limited email-change and password-reset tokens;
- HTTP-only, `SameSite=Lax` frontend cookies;
- secure-cookie mode in production;
- Cloudflare Turnstile validation;
- optional TOTP two-factor authentication;
- encrypted TOTP secrets;
- hashed one-time recovery codes;
- short-lived, single-use 2FA challenges.

### Portal security

Portal accounts use their own Argon2id password hashes and random, hashed session tokens. Portal sessions record request-source and user-agent information. Account type and `account_enabled` status are checked during protected portal access. Course, post, assignment, and submission handlers also confirm that the student is actually assigned to the relevant course or owns the relevant submission.

### API safeguards

The Go API uses Chi middleware for request IDs, logging, panic recovery, client IP handling, heartbeat checks, and a 15-second request timeout. JSON handlers generally reject unknown fields and distinguish malformed input, validation failures, conflicts, unauthenticated access, forbidden access, and server errors through HTTP status codes and targeted error codes.

Destructive operations use PostgreSQL transactions where several related changes must succeed together. Database foreign keys, unique indexes, check constraints, partial indexes, and cascades reinforce application-level validation.

## File-storage architecture

EduLink uses S3-compatible object storage through the MinIO Go client. Staff and portal uploads use separate object metadata tables so ownership remains explicit.

The upload lifecycle is:

1. The authenticated user requests an upload initialization.
2. The API validates ownership, permissions, filename, declared size, and content type.
3. The API creates a pending storage-object record and completion token.
4. The frontend uploads directly to a presigned object-storage URL.
5. The frontend calls the completion endpoint.
6. The API verifies the stored object and marks the record done or failed.
7. Completed object records are joined to a post, profile picture, or submission attachment.
8. Read handlers generate temporary presigned URLs instead of exposing the storage bucket publicly.

The frontend validates files before upload and shows storage-specific errors without treating client validation as authoritative. Object-storage CORS must allow the deployed frontend origin for browser-to-storage uploads and downloads.

## Email delivery

EduLink uses SMTP for transactional staff mail, including registration, password reset, email verification, and staff invitations. Invitation messages can carry an email importance selected by the inviter.

Bulk student account creation writes personalized account emails into a PostgreSQL-backed queue. Queue records include recipient, subject, body, content type, priority, purpose, scheduled time, processing status, retry limit, retry count, and last retry time.

The repository contains queue creation and schema support, but no active queue-processing worker is visible in the current API startup path. A deployment therefore needs a worker that claims pending messages, sends them, updates retry state, and marks them sent or failed before queued student credentials will actually leave the system.

## Data model

The principal database relationships are:

```text
Staff user
├── staff sessions
├── verification and registration flows
├── two-factor challenges and recovery codes
├── owned schools
├── school staff memberships
├── role memberships
├── authored posts and grades
└── profile picture storage

School
├── staff memberships and invitations
├── ordered roles and permissions
├── academic years
│   └── grades
│       └── courses
│           ├── assigned students
│           ├── posts
│           │   └── post attachments
│           └── assignments
│               └── submissions
│                   ├── submission attachments
│                   └── score and teacher feedback
├── portal users
└── audit logs

Portal user
├── portal sessions
├── assigned courses
└── assignment submissions
```

Most primary IDs are generated as 64-bit Snowflake-style values. The API returns them to JavaScript as strings to avoid precision loss. A database-backed machine-ID lease allows multiple API instances to generate IDs without sharing the same worker identifier.

The database uses a reserved deleted-user record for historical content whose staff author has been removed. This preserves post, grading, and audit references without retaining a live user relationship.

## Technical stack

### Frontend

- Next.js 16 App Router
- React 19
- strict TypeScript
- Tailwind CSS 4
- Base UI and shadcn-based reusable components
- Framer Motion and GSAP for motion
- Sonner for toast feedback
- Lucide icons
- QR code generation for TOTP setup
- server components and server actions for authenticated API access

### Backend

- Go 1.25
- Chi router and middleware
- PostgreSQL through `database/sql` and `lib/pq`
- Argon2id password hashing
- SHA-256 token hashing
- TOTP and QR-compatible two-factor setup
- Sonyflake-compatible distributed IDs with database-leased machine IDs
- MinIO client for S3-compatible storage
- SMTP email delivery
- PostgreSQL migrations for schema evolution

### Deployment support

- a multi-stage backend Dockerfile;
- Docker Compose configuration for the API container;
- environment-driven database, storage, SMTP, frontend-origin, encryption, and application configuration;
- separate production build, lint, TypeScript, and test scripts for the frontend.

## Current boundaries

The following areas are not complete product features yet:

- The guardian portal is a routed placeholder without guardian-specific data or workflows.
- Students cannot directly edit school-managed profile information.
- Student profile pictures are not implemented.
- Audit-log pagination is frontend-side; the API still returns the full accessible log collection.
- Audit-log content itself is not localized because it is stored as backend-generated English text.
- The database email queue needs an operational worker outside the startup code currently present in this repository.
- Assignment notification subscriptions were explored in migrations and later removed; assignment notifications are not a current feature.
- Student archiving is not exposed as a frontend or current data-model feature.
- The project does not currently include messaging, attendance, schedules, guardian-to-student relationships, billing, or analytics dashboards.

These boundaries are intentional to the current scope. The course dashboard, portal account model, permission catalog, and separate staff/portal route trees are structured so additional school and guardian features can be added without replacing the existing core workflows.
