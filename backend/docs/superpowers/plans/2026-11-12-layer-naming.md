# Application and HTTP naming cleanup

Goal: Improve package, file, handler, and route registration names while preserving the architecture and HTTP contract.

Architecture: Keep application services, HTTP handlers, and chi route packages separate. Keep the staff/portal boundaries and service boundaries.

Constraints: Preserve URLs, methods, route parameters, middleware, payloads, and handler logic. Do not add tests. Leave the existing README edits alone. Execute directly in the current workspace under the user’s broad authorization; do not commit.

- [ ] Capture the existing chi route table and build baseline.
- [ ] Rename staff courseposts to posts, coursepostattachments to attachments, coursestudents to enrollments, and submissionscores to scores; update imports and wiring. Keep academic years and portal users descriptive rather than inventing vague shorter names.
- [ ] Remove redundant Handler method suffixes and replace abbreviations/noun-first operation names. Align application operation filenames and HTTP resource filenames.
- [ ] Give staff and portal route packages a Routes entry point, keep resource helpers private, consolidate tiny helpers with their resource, and group school/course collections in their owning resource files. Preserve mount semantics.
- [ ] Compare route/middleware/handler inventory, build, vet, and inspect semantic diffs. No new tests.

Review focus: Nested mount slash behavior, route parameter names, static routes vs parameter routes, school role route reuse, and middleware order.
