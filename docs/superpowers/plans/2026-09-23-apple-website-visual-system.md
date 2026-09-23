# Apple.com-Inspired Frontend Visual System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restyle the complete EduLink frontend to closely follow Apple.com’s restrained visual language in light and dark modes without changing product behavior or information architecture.

**Architecture:** Replace the route-specific students experiment with a semantic shared design system. Establish the palette and typography in global tokens, update existing UI primitives at their source, then remove conflicting decorative styling from application shells and route components while preserving every API, state transition, permission check, and responsive layout.

**Tech Stack:** Next.js 16 App Router, React 19, TypeScript, Tailwind CSS 4, Base UI, class-variance-authority, Lucide React, Sonner.

**Spec:** `docs/superpowers/specs/2026-09-23-apple-website-visual-system-design.md`

## Global Constraints

- Do not modify anything in `backend/`.
- Preserve all routes, workflows, permissions, server actions, validation, translations, and data behavior.
- Keep light mode as the default and preserve the existing light/dark theme switch and cookie behavior.
- Use US English locale conventions; do not introduce British locales or formatting.
- Do not add dependencies unless a hard requirement is discovered and approved.
- Do not use Apple logos, product imagery, copied marketing text, or other Apple brand assets.
- Imitate Apple.com’s website language, not iOS or macOS application chrome.
- Avoid decorative gradients, glowing shapes, grid textures, heavy blur, exaggerated glass effects, and card-per-section layouts.
- Preserve existing user changes in the dirty worktree. Stage and commit only the explicit files owned by each task.
- Do not stage the existing course action/page changes or unrelated backend and documentation changes.

## Review Focus

- Theme preference: switching light and dark must still update the document theme and persist across navigation without layout changes.
- Destructive actions: irreversible controls must remain visibly destructive in both themes, with confirmation and disabled states unchanged.
- Submission safety: loading and disabled buttons must remain readable and prevent duplicate interaction.
- Dense data: tables, import grids, audit logs, and metadata rows must remain scrollable and readable on narrow screens.
- Long localized copy: English and Polish labels, descriptions, navigation items, and empty states must wrap without overlapping controls.

---

## File Structure

- `frontend/src/app/globals.css`: semantic palette, font stack, radii, focus treatment, and reduced-motion defaults.
- `frontend/src/components/ui/*.tsx`: shared controls, content surfaces, overlays, data display, and state presentation.
- `frontend/src/components/app/*.tsx`: shared application presentation.
- `frontend/src/components/auth/*.tsx`: shared authentication presentation.
- `frontend/src/components/landing/*.tsx`: public marketing presentation.
- `frontend/src/app/app/client_layout.tsx`: staff workspace shell.
- `frontend/src/app/app/(protected)/staff/(school_navbar)/layout.tsx`: school navigation shell.
- `frontend/src/app/app/portal/portal_shell.tsx`: authenticated portal shell.
- Route client/loading files: remove only styling that conflicts with shared primitives.

---

### Task 1: Remove the Route-Scoped Students Prototype

**Files:**
- Delete: `frontend/src/lib/apple_students_theme.ts`
- Delete: `frontend/src/lib/apple_students_theme.test.mjs`
- Modify: `frontend/src/app/app/client_layout.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/layout.tsx`
- Modify: `frontend/src/app/globals.css`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/client_page.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/loading.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/student_form.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/bulk_student_import.tsx`

**Interfaces:**
- Consumes: existing layout shells and shared UI primitives.
- Produces: no route-scoped theme API; students return to shared semantic styling.

- [ ] **Step 1: Record the behavior baseline**

```bash
cd frontend
npm test
npx tsc --noEmit
```

Expected: all current tests pass and TypeScript exits successfully.

- [ ] **Step 2: Remove the route detector and shell branches**

Delete both `apple_students_theme` files. In `client_layout.tsx`, remove `usePathname`, `cn`, `isAppleStudentsThemeRoute`, the derived boolean, and the conditional class. Restore the root to:

```tsx
<motion.div
    initial={{opacity: 0}}
    animate={{opacity: 1}}
    exit={{opacity: 0}}
>
```

In the school layout, remove the helper import, derived boolean, and `apple-liquid-mobile-nav` branch.

- [ ] **Step 3: Remove the prototype CSS block**

Delete the complete block beginning with:

```css
/* Route-scoped visual prototype for the school students index. */
```

Do not alter the original semantic theme variables in this step.

- [ ] **Step 4: Normalize students markup**

Remove `apple-*` classes from the four students files. Retain responsive title sizing, the summary grid, accessible labels, mobile stacking, destructive confirmation, and existing behavior. Dialog sizing remains explicit:

```tsx
<DialogPopup className="sm:max-w-2xl">
<DialogPopup className="sm:max-w-6xl">
<AlertDialogPopup className="sm:max-w-xl">
```

- [ ] **Step 5: Verify removal**

```bash
cd frontend
rg -n "apple-liquid|apple-students|apple-summary|apple-status|isAppleStudentsThemeRoute" src
npm test
npx tsc --noEmit
```

Expected: `rg` finds no matches; tests and TypeScript pass.

- [ ] **Step 6: Commit only prototype-removal files**

```bash
git add -- frontend/src/app/globals.css frontend/src/app/app/client_layout.tsx 'frontend/src/app/app/(protected)/staff/(school_navbar)/layout.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/client_page.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/loading.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/student_form.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/bulk_student_import.tsx'
git commit -m "refactor: remove students theme prototype"
```

### Task 2: Establish the Visual Foundation

**Files:**
- Modify: `frontend/src/app/globals.css`
- Modify: `frontend/src/app/layout.tsx`
- Modify: `frontend/src/components/app/page_title.tsx`
- Modify: `frontend/src/components/ui/button.tsx`
- Modify: `frontend/src/components/ui/card.tsx`
- Modify: `frontend/src/components/ui/badge.tsx`
- Modify: `frontend/src/components/ui/separator.tsx`
- Modify: `frontend/src/components/ui/avatar.tsx`

**Interfaces:**
- Consumes: existing semantic CSS variable names.
- Produces: stable tokens and base primitives used by every later task.

- [ ] **Step 1: Capture light and dark baselines**

```bash
cd frontend
npm run dev
```

Capture the landing page, staff login, and one authenticated page at 375px and 1440px in both themes. If authentication is unavailable, capture public pages and list authenticated visual QA as an explicit final limitation.

- [ ] **Step 2: Replace semantic theme values**

Keep existing variable names. Set the light theme core to:

```css
:root {
  color-scheme: light;
  --background: #ffffff;
  --foreground: #1d1d1f;
  --card: #ffffff;
  --card-foreground: #1d1d1f;
  --popover: #ffffff;
  --popover-foreground: #1d1d1f;
  --primary: #0071e3;
  --primary-foreground: #ffffff;
  --secondary: #f5f5f7;
  --secondary-foreground: #1d1d1f;
  --muted: #f5f5f7;
  --muted-foreground: #6e6e73;
  --accent: #f5f5f7;
  --accent-foreground: #1d1d1f;
  --destructive: #d70015;
  --border: rgb(0 0 0 / 10%);
  --input: rgb(0 0 0 / 14%);
  --ring: #0071e3;
  --radius: 0.75rem;
}
```

Set dark mode to black/near-black, `#1d1d1f`, `#f5f5f7`, `#a1a1a6`, low-opacity white separators, and `#2997ff`. Keep status, chart, and sidebar variables semantically distinct.

- [ ] **Step 3: Make the system font authoritative**

```css
body {
  font-family: -apple-system, BlinkMacSystemFont, "SF Pro Text", "SF Pro Display", Inter, Arial, Helvetica, sans-serif;
  letter-spacing: -0.011em;
}
```

Keep `lang={locale}` in the root layout and change the top loader accent to `#0071e3`.

- [ ] **Step 4: Restyle headings and actions**

Use semibold page titles with negative tracking and responsive size. Keep all button variants and props. Default primary actions are blue pills; outline/secondary are neutral pills; destructive remains red; icon-only controls are circular.

```tsx
"relative inline-flex shrink-0 cursor-pointer items-center justify-center gap-2 whitespace-nowrap border font-medium outline-none transition-[color,background-color,border-color,box-shadow,transform] focus-visible:ring-2 focus-visible:ring-ring/30 disabled:pointer-events-none disabled:opacity-50"
```

- [ ] **Step 5: Restyle static surfaces**

Remove default card inset highlights and shadows. Standard cards use `rounded-2xl border bg-card`. Update badges, separators, and avatars to use semantic colors and subtle borders.

- [ ] **Step 6: Test foundation failure modes**

Verify theme persistence across navigation, stable loading-button width, duplicate-click prevention, red destructive actions, and keyboard focus in both themes.

```bash
cd frontend
npx eslint src/app/layout.tsx src/components/app/page_title.tsx src/components/ui/button.tsx src/components/ui/card.tsx src/components/ui/badge.tsx src/components/ui/separator.tsx src/components/ui/avatar.tsx
npx tsc --noEmit
```

- [ ] **Step 7: Commit foundation files**

```bash
git add -- frontend/src/app/globals.css frontend/src/app/layout.tsx frontend/src/components/app/page_title.tsx frontend/src/components/ui/button.tsx frontend/src/components/ui/card.tsx frontend/src/components/ui/badge.tsx frontend/src/components/ui/separator.tsx frontend/src/components/ui/avatar.tsx
git commit -m "style: establish Apple-inspired frontend foundation"
```

### Task 3: Restyle Form and Selection Controls

**Files:**
- Modify: `frontend/src/components/ui/input.tsx`
- Modify: `frontend/src/components/ui/textarea.tsx`
- Modify: `frontend/src/components/ui/select.tsx`
- Modify: `frontend/src/components/ui/autocomplete.tsx`
- Modify: `frontend/src/components/ui/combobox.tsx`
- Modify: `frontend/src/components/ui/field.tsx`
- Modify: `frontend/src/components/ui/fieldset.tsx`
- Modify: `frontend/src/components/ui/label.tsx`
- Modify: `frontend/src/components/ui/checkbox.tsx`
- Modify: `frontend/src/components/ui/checkbox-group.tsx`
- Modify: `frontend/src/components/ui/radio-group.tsx`
- Modify: `frontend/src/components/ui/switch.tsx`
- Modify: `frontend/src/components/ui/toggle.tsx`
- Modify: `frontend/src/components/ui/toggle-group.tsx`
- Modify: `frontend/src/components/ui/number-field.tsx`
- Modify: `frontend/src/components/ui/otp-field.tsx`
- Modify: `frontend/src/components/ui/slider.tsx`
- Modify: `frontend/src/components/ui/input-group.tsx`
- Modify: `frontend/src/components/ui/form.tsx`
- Modify: `frontend/src/components/ui/calendar.tsx`

**Interfaces:**
- Consumes: Task 2 tokens, radii, focus ring, and buttons.
- Produces: consistent form controls with unchanged Base UI and React props.

- [ ] **Step 1: Apply one form-control contract**

```tsx
"relative inline-flex w-full rounded-xl border border-input bg-background text-foreground shadow-none ring-ring/20 transition-[border-color,box-shadow,background-color] has-focus-visible:border-ring has-focus-visible:ring-[3px] has-disabled:bg-muted/50 has-disabled:opacity-60"
```

Use a 44px minimum height on touch contexts and at least 36px on desktop. Keep invalid attributes and component slots intact.

- [ ] **Step 2: Restyle text entry**

Apply the contract to input, textarea, number, OTP, and input-group components. Preserve native types, autocomplete, disabled behavior, and Base UI render plumbing.

- [ ] **Step 3: Restyle selection controls**

Use simple checkmarks, circular radio indicators, and a system-like switch. Checked switches use primary blue. Toggles use neutral selected fills.

```tsx
"inline-flex shrink-0 items-center rounded-full bg-input p-0.5 outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring/30 data-checked:bg-primary data-disabled:opacity-50"
```

- [ ] **Step 4: Restyle fields and calendar**

Keep labels sentence case and semibold, descriptions secondary, and errors red. Calendar selection uses a blue filled day with no decorative shadow.

- [ ] **Step 5: Test form failure modes**

On login, registration, student create/edit, and settings, verify keyboard order, invalid styling, disabled legibility, duplicate-submit prevention, and long Polish labels at 320px.

```bash
cd frontend
npx eslint src/components/ui
npx tsc --noEmit
```

- [ ] **Step 6: Commit form files**

```bash
git add -- frontend/src/components/ui/input.tsx frontend/src/components/ui/textarea.tsx frontend/src/components/ui/select.tsx frontend/src/components/ui/autocomplete.tsx frontend/src/components/ui/combobox.tsx frontend/src/components/ui/field.tsx frontend/src/components/ui/fieldset.tsx frontend/src/components/ui/label.tsx frontend/src/components/ui/checkbox.tsx frontend/src/components/ui/checkbox-group.tsx frontend/src/components/ui/radio-group.tsx frontend/src/components/ui/switch.tsx frontend/src/components/ui/toggle.tsx frontend/src/components/ui/toggle-group.tsx frontend/src/components/ui/number-field.tsx frontend/src/components/ui/otp-field.tsx frontend/src/components/ui/slider.tsx frontend/src/components/ui/input-group.tsx frontend/src/components/ui/form.tsx frontend/src/components/ui/calendar.tsx
git commit -m "style: simplify frontend form controls"
```

### Task 4: Restyle Dialogs and Floating Layers

**Files:**
- Modify: `frontend/src/components/ui/dialog.tsx`
- Modify: `frontend/src/components/ui/alert-dialog.tsx`
- Modify: `frontend/src/components/ui/sheet.tsx`
- Modify: `frontend/src/components/ui/drawer.tsx`
- Modify: `frontend/src/components/ui/menu.tsx`
- Modify: `frontend/src/components/ui/context-menu.tsx`
- Modify: `frontend/src/components/ui/popover.tsx`
- Modify: `frontend/src/components/ui/preview-card.tsx`
- Modify: `frontend/src/components/ui/tooltip.tsx`
- Modify: `frontend/src/components/ui/command.tsx`
- Modify: `frontend/src/components/ui/scroll-area.tsx`
- Modify: `frontend/src/components/ui/accordion.tsx`
- Modify: `frontend/src/components/ui/collapsible.tsx`
- Modify: `frontend/src/components/ui/toast.tsx`
- Modify: `frontend/src/components/app/global_context_menu.tsx`

**Interfaces:**
- Consumes: Task 2 cards/buttons and Task 3 form controls.
- Produces: shared floating-layer styling for dialogs, mobile navigation, menus, and feedback.

- [ ] **Step 1: Apply one floating-surface contract**

Use opaque or nearly opaque surfaces, one border, moderate radii, and one soft shadow:

```tsx
"rounded-2xl border border-border bg-popover text-popover-foreground shadow-[0_18px_50px_rgb(0_0_0/0.14)] outline-none"
```

Backdrop blur is permitted on modal backdrops or sticky navigation only.

- [ ] **Step 2: Restyle modal containers**

Apply the contract to dialogs, alert dialogs, sheets, and drawers. Preserve portals, nested-dialog variables, mobile positioning, focus traps, close behavior, and scrolling.

- [ ] **Step 3: Restyle menus and transient surfaces**

Use compact spacing and clear selected, destructive, and disabled states for menus, context menus, popovers, preview cards, tooltips, and command palettes. Remove inset highlights and glass effects.

- [ ] **Step 4: Restyle disclosure and feedback**

Use separator-driven accordions and collapsibles. Keep toast semantics and animation while simplifying surfaces and shadows.

- [ ] **Step 5: Test overlay failure modes**

Verify dialog focus trapping and close behavior, destructive confirmation gating, mobile sheet viewport fit, long English/Polish descriptions, and menu keyboard navigation in both themes.

```bash
cd frontend
npx eslint src/components/ui/dialog.tsx src/components/ui/alert-dialog.tsx src/components/ui/sheet.tsx src/components/ui/drawer.tsx src/components/ui/menu.tsx src/components/ui/context-menu.tsx src/components/ui/popover.tsx src/components/ui/preview-card.tsx src/components/ui/tooltip.tsx src/components/ui/command.tsx src/components/ui/scroll-area.tsx src/components/ui/accordion.tsx src/components/ui/collapsible.tsx src/components/ui/toast.tsx src/components/app/global_context_menu.tsx
npx tsc --noEmit
```

- [ ] **Step 6: Commit floating layers**

```bash
git add -- frontend/src/components/ui/dialog.tsx frontend/src/components/ui/alert-dialog.tsx frontend/src/components/ui/sheet.tsx frontend/src/components/ui/drawer.tsx frontend/src/components/ui/menu.tsx frontend/src/components/ui/context-menu.tsx frontend/src/components/ui/popover.tsx frontend/src/components/ui/preview-card.tsx frontend/src/components/ui/tooltip.tsx frontend/src/components/ui/command.tsx frontend/src/components/ui/scroll-area.tsx frontend/src/components/ui/accordion.tsx frontend/src/components/ui/collapsible.tsx frontend/src/components/ui/toast.tsx frontend/src/components/app/global_context_menu.tsx
git commit -m "style: simplify dialogs and floating layers"
```

### Task 5: Restyle Data, Navigation, and State Primitives

**Files:**
- Modify: `frontend/src/components/ui/table.tsx`
- Modify: `frontend/src/components/ui/tabs.tsx`
- Modify: `frontend/src/components/ui/pagination.tsx`
- Modify: `frontend/src/components/ui/breadcrumb.tsx`
- Modify: `frontend/src/components/ui/toolbar.tsx`
- Modify: `frontend/src/components/ui/group.tsx`
- Modify: `frontend/src/components/ui/kbd.tsx`
- Modify: `frontend/src/components/ui/meter.tsx`
- Modify: `frontend/src/components/ui/progress.tsx`
- Modify: `frontend/src/components/ui/skeleton.tsx`
- Modify: `frontend/src/components/ui/spinner.tsx`
- Modify: `frontend/src/components/ui/empty.tsx`
- Modify: `frontend/src/components/ui/alert.tsx`
- Modify: `frontend/src/components/ui/frame.tsx`
- Modify: `frontend/src/components/ui/sidebar.tsx`
- Modify: `frontend/src/components/app/post-attachments.tsx`

**Interfaces:**
- Consumes: semantic colors and controls from Tasks 2–4.
- Produces: consistent dense-data, navigation, and state presentation.

- [ ] **Step 1: Restyle tables and tabs**

Use thin separators and row hover fills instead of card-shaped rows. Keep horizontal overflow on the table container. Use a neutral segmented fill for default tabs and a two-pixel blue indicator for underline tabs.

```tsx
"relative border-b border-border transition-colors hover:bg-muted/45 data-[state=selected]:bg-muted"
```

- [ ] **Step 2: Restyle navigation helpers**

Use plain text and chevrons for breadcrumbs and pagination. Toolbars and groups receive a neutral background only when grouping improves comprehension.

- [ ] **Step 3: Restyle state communication**

Use restrained tinted fills for alerts, compact empty states, a low-contrast skeleton shimmer, and simple progress/meter tracks. Retain text/icons so state never relies only on color.

- [ ] **Step 4: Restyle frames and attachments**

Remove inset highlights and nested elevation from frames, the unused sidebar primitive, and post attachments. Preserve download, preview, and removal behavior.

- [ ] **Step 5: Test dense-data failure modes**

At 375px and desktop widths, verify table horizontal scrolling, final action-column access, long Polish tab labels, pagination disabled states, all alert variants, empty/loading states, and attachment keyboard focus.

```bash
cd frontend
npx eslint src/components/ui/table.tsx src/components/ui/tabs.tsx src/components/ui/pagination.tsx src/components/ui/breadcrumb.tsx src/components/ui/toolbar.tsx src/components/ui/group.tsx src/components/ui/kbd.tsx src/components/ui/meter.tsx src/components/ui/progress.tsx src/components/ui/skeleton.tsx src/components/ui/spinner.tsx src/components/ui/empty.tsx src/components/ui/alert.tsx src/components/ui/frame.tsx src/components/ui/sidebar.tsx src/components/app/post-attachments.tsx
npx tsc --noEmit
```

- [ ] **Step 6: Commit data and state primitives**

```bash
git add -- frontend/src/components/ui/table.tsx frontend/src/components/ui/tabs.tsx frontend/src/components/ui/pagination.tsx frontend/src/components/ui/breadcrumb.tsx frontend/src/components/ui/toolbar.tsx frontend/src/components/ui/group.tsx frontend/src/components/ui/kbd.tsx frontend/src/components/ui/meter.tsx frontend/src/components/ui/progress.tsx frontend/src/components/ui/skeleton.tsx frontend/src/components/ui/spinner.tsx frontend/src/components/ui/empty.tsx frontend/src/components/ui/alert.tsx frontend/src/components/ui/frame.tsx frontend/src/components/ui/sidebar.tsx frontend/src/components/app/post-attachments.tsx
git commit -m "style: refine data and state presentation"
```

### Task 6: Harmonize Public and Authenticated Shells

**Files:**
- Modify: `frontend/src/components/landing/landing-page.tsx`
- Modify: `frontend/src/components/landing/app-demos.tsx`
- Modify: `frontend/src/components/auth/auth-page-shell.tsx`
- Modify: `frontend/src/components/auth/auth-card.tsx`
- Modify: `frontend/src/components/app/language-switcher.tsx`
- Modify: `frontend/src/app/app/client_layout.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/layout.tsx`
- Modify: `frontend/src/app/app/portal/client_page.tsx`
- Modify: `frontend/src/app/app/portal/portal_shell.tsx`
- Modify: `frontend/src/components/legal/legal-document.tsx`
- Modify: `frontend/src/app/legal/layout.tsx`

**Interfaces:**
- Consumes: the shared visual system from Tasks 2–5.
- Produces: consistent public, staff, portal, auth, and legal shells.

- [ ] **Step 1: Harmonize landing and legal shells**

Replace hard-coded light-only colors with semantic tokens while preserving the landing page’s existing Apple.com-like structure, 48px navigation, blue actions, full-width sections, and type hierarchy. Keep the legal sidebar and document structure.

- [ ] **Step 2: Remove decorative auth backgrounds**

Delete radial-gradient and grid overlays from `auth-page-shell.tsx`. Use a solid primary background and a centered flat auth card. Remove the auth card’s oversized shadow and decorative header fill while retaining slots and content.

- [ ] **Step 3: Simplify the staff shell**

Replace the diagonal gradient in `client_layout.tsx` with `bg-background`. Make the footer quiet. Use a soft-gray/graphite school sidebar, thin separators, and a blue active label/indicator without raised-card effects. Preserve sticky and mobile behavior.

- [ ] **Step 4: Simplify portal shells**

Remove portal radial backgrounds and glass feature cards. Keep the two-column login layout with flat soft-gray sections. Align authenticated portal navigation with the compact staff hierarchy while preserving route logic and theme switching.

- [ ] **Step 5: Test shell failure modes**

Verify theme persistence, destination parity between desktop/mobile navigation, active states, long Polish labels, 320px auth-card fit, and landing/legal readability under a saved dark theme.

```bash
cd frontend
npx eslint src/components/landing/landing-page.tsx src/components/landing/app-demos.tsx src/components/auth/auth-page-shell.tsx src/components/auth/auth-card.tsx src/components/app/language-switcher.tsx src/app/app/client_layout.tsx 'src/app/app/(protected)/staff/(school_navbar)/layout.tsx' src/app/app/portal/client_page.tsx src/app/app/portal/portal_shell.tsx src/components/legal/legal-document.tsx src/app/legal/layout.tsx
npx tsc --noEmit
```

- [ ] **Step 6: Commit shell changes**

```bash
git add -- frontend/src/components/landing/landing-page.tsx frontend/src/components/landing/app-demos.tsx frontend/src/components/auth/auth-page-shell.tsx frontend/src/components/auth/auth-card.tsx frontend/src/components/app/language-switcher.tsx frontend/src/app/app/client_layout.tsx 'frontend/src/app/app/(protected)/staff/(school_navbar)/layout.tsx' frontend/src/app/app/portal/client_page.tsx frontend/src/app/app/portal/portal_shell.tsx frontend/src/components/legal/legal-document.tsx frontend/src/app/legal/layout.tsx
git commit -m "style: harmonize frontend application shells"
```

### Task 7: Clean Up Staff Route Presentation

**Files:**
- Modify: `frontend/src/app/app/(protected)/client_page.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/client_page.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/staff/client_page.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/roles/client_page.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/roles/loading.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/settings/client_page.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/logs/client_page.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/client_page.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/loading.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/student_form.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/bulk_student_import.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/students/[studentID]/client_page.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/grades/[gradeID]/client_page.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/courses/[courseID]/client_page.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/courses/[courseID]/assignments_section.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/courses/[courseID]/post_attachment_picker.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/assignments/[assignmentID]/submissions/client_page.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/(school_navbar)/submissions/[submissionID]/client_page.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/profile/client_page.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/profile/two_factor_setup.tsx`
- Modify: `frontend/src/app/app/(protected)/staff/posts/[postID]/back_button.tsx`

**Interfaces:**
- Consumes: all shared primitives and shells.
- Produces: staff routes free of conflicting decorative styling with behavior unchanged.

- [ ] **Step 1: Locate route-level overrides**

```bash
cd frontend
rg -n "gradient|backdrop-blur|shadow-xl|shadow-lg|rounded-3xl|bg-card/|bg-muted/" 'src/app/app/(protected)'
```

Remove decorative classes only. Preserve grids, widths, overflow, sticky positioning, functional spacing, and semantic status colors.

- [ ] **Step 2: Reduce unnecessary card nesting**

Where one card contains only another card or a bordered row group, retain one surface and use separators inside it. Do not move server/client boundaries, state, or data transformations.

- [ ] **Step 3: Normalize headers and actions**

Use `PageTitle` or its class contract. Keep primary actions blue, secondary actions neutral, and destructive actions red. Preserve all permission-driven conditional rendering.

- [ ] **Step 4: Normalize loading, empty, and error geometry**

Make skeletons match final control and row sizes. Preserve every branch and translation key. Empty states remain concise without decorative card stacks or oversized icon tiles.

- [ ] **Step 5: Test staff workflows**

Verify school search/invitations/creation; student search/pagination/create/edit/import/delete; permission-hidden actions; staff/role/log/grade/course/assignment/submission tables at 375px; and loading-button duplicate prevention. Confirm the existing user edits in the course route remain present and do not edit its `actions.ts`.

```bash
cd frontend
npm test
npx eslint 'src/app/app/(protected)'
npx tsc --noEmit
```

- [ ] **Step 6: Commit only reviewed staff files**

Inspect `git diff -- <path>` before staging every file. Exclude any diff containing unrelated user work, especially the existing course `actions.ts` and `page.tsx` changes.

```bash
git add -- 'frontend/src/app/app/(protected)/client_page.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/client_page.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/staff/client_page.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/roles/client_page.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/roles/loading.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/settings/client_page.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/logs/client_page.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/client_page.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/loading.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/student_form.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/schools/[id]/students/bulk_student_import.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/students/[studentID]/client_page.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/grades/[gradeID]/client_page.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/courses/[courseID]/client_page.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/courses/[courseID]/assignments_section.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/courses/[courseID]/post_attachment_picker.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/assignments/[assignmentID]/submissions/client_page.tsx' 'frontend/src/app/app/(protected)/staff/(school_navbar)/submissions/[submissionID]/client_page.tsx' 'frontend/src/app/app/(protected)/staff/profile/client_page.tsx' 'frontend/src/app/app/(protected)/staff/profile/two_factor_setup.tsx' 'frontend/src/app/app/(protected)/staff/posts/[postID]/back_button.tsx'
git commit -m "style: align staff routes with shared visual system"
```

### Task 8: Clean Up Portal, Invitation, and Auth Routes

**Files:**
- Modify: `frontend/src/app/app/portal/auth/activate/[token]/client_page.tsx`
- Modify: `frontend/src/app/app/portal/auth/activate/[token]/loading.tsx`
- Modify: `frontend/src/app/app/portal/loading.tsx`
- Modify: `frontend/src/app/app/portal/student/client_page.tsx`
- Modify: `frontend/src/app/app/portal/student/loading.tsx`
- Modify: `frontend/src/app/app/portal/student/assignments/page.tsx`
- Modify: `frontend/src/app/app/portal/student/assignments/loading.tsx`
- Modify: `frontend/src/app/app/portal/student/courses/[courseID]/page.tsx`
- Modify: `frontend/src/app/app/portal/student/courses/[courseID]/loading.tsx`
- Modify: `frontend/src/app/app/portal/student/courses/[courseID]/assignments_section.tsx`
- Modify: `frontend/src/app/app/portal/student/courses/[courseID]/submission_dialog.tsx`
- Modify: `frontend/src/app/app/portal/student/profile/page.tsx`
- Modify: `frontend/src/app/app/portal/student/profile/loading.tsx`
- Modify: `frontend/src/app/app/staff-invitations/[token]/client_page.tsx`
- Modify: `frontend/src/app/app/staff-invitations/[token]/loading.tsx`
- Modify: `frontend/src/app/auth/(protected)/login/page.tsx`
- Modify: `frontend/src/app/auth/loading.tsx`
- Modify: `frontend/src/app/auth/verify-change/[token]/page.tsx`

**Interfaces:**
- Consumes: shared primitives and portal/auth shells.
- Produces: remaining frontend flows aligned with the shared visual system.

- [ ] **Step 1: Locate remaining decorative styles**

```bash
cd frontend
rg -n "gradient|backdrop-blur|shadow-xl|shadow-lg|rounded-3xl|bg-card/|bg-muted/" src/app/app/portal src/app/app/staff-invitations src/app/auth
```

Remove decorative treatments while preserving responsive grids, status messaging, form structure, and route layout.

- [ ] **Step 2: Normalize content hierarchy**

Use shared titles, cards, forms, alerts, and skeletons. Keep coursework and assignment pages dense enough for scanning; do not turn detail pages into marketing heroes.

- [ ] **Step 3: Test auth and portal workflows**

Verify login/activation error announcements, verification configured/missing branches, portal navigation, course links, submission dialogs, profile actions, long labels at 320px, and dark-mode disabled/status contrast.

```bash
cd frontend
npm test
npx eslint src/app/app/portal src/app/app/staff-invitations src/app/auth
npx tsc --noEmit
```

- [ ] **Step 4: Commit portal and auth cleanup**

```bash
git add -- frontend/src/app/app/portal frontend/src/app/app/staff-invitations frontend/src/app/auth
git commit -m "style: align portal and auth routes"
```

### Task 9: Final Cross-Route Verification

**Files:**
- Modify only files required to fix issues found by the checks below.

**Interfaces:**
- Consumes: the complete visual-system implementation.
- Produces: verified production-ready frontend restyle.

- [ ] **Step 1: Scan for conflicting visual language**

```bash
cd frontend
rg -n "radial-gradient|linear-gradient|backdrop-blur|shadow-xl|shadow-lg|apple-liquid|apple-students" src/app src/components
```

Expected: only sticky-navigation translucency, modal backdrops, skeleton shimmer, or functional artwork remains. Inspect every match and remove decorative leftovers.

- [ ] **Step 2: Run the complete automated suite**

```bash
cd frontend
npm test
npm run lint
npx tsc --noEmit
npm run build
```

Expected: every command exits with status 0. Report framework deprecation warnings separately from failures.

- [ ] **Step 3: Run repository hygiene checks**

```bash
git diff --check
git status --short
git diff --stat -- frontend
```

Expected: no whitespace errors, no backend file included in the redesign, and unrelated changes remain intact.

- [ ] **Step 4: Perform the visual matrix**

Inspect these routes in light/dark at 375px and 1440px:

```text
/
/auth/login
/app
/app/staff/schools/{schoolID}
/app/staff/schools/{schoolID}/students
/app/staff/schools/{schoolID}/staff
/app/staff/schools/{schoolID}/roles
/app/staff/schools/{schoolID}/logs
/app/staff/schools/{schoolID}/settings
/app/staff/courses/{courseID}
/app/portal
/app/portal/student
/legal/privacy
```

Check navigation, long text, focus, hover, loading, empty, error, dialog, destructive, table overflow, and disabled states. If browser access or authentication blocks a route, list the exact unverified route in the handoff.

- [ ] **Step 5: Commit verification fixes**

If verification finds issues, stage each exact fix path after reviewing its diff and confirm the staged list before committing. If no fixes are needed, skip this commit.

```bash
git diff --cached --name-only
git commit -m "fix: polish Apple-inspired frontend restyle"
```

- [ ] **Step 6: Request final code review**

Review the full frontend diff against the design spec, focusing on behavior changes, contrast, dark mode, mobile overflow, destructive actions, and accidental inclusion of user-owned changes.
