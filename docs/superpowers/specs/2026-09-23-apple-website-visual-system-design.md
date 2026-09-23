# Apple.com-Inspired Frontend Visual System

## Purpose

Restyle the entire EduLink frontend so it closely follows the visual language of Apple.com while preserving the product's existing layouts, workflows, permissions, routes, and data behavior.

The redesign must feel intentional and restrained rather than decorative. Light mode remains the default, and dark mode remains available through the existing theme switch.

## Design Target

The target is Apple's marketing website, not an imitation of iOS, macOS, or a concept dashboard.

Core characteristics:

- System-font typography with strong hierarchy, tight heading tracking, and highly legible body copy.
- Predominantly white and soft-gray surfaces in light mode.
- Predominantly black and graphite surfaces in dark mode.
- Apple blue reserved for links, focus, selection, and primary actions.
- Generous whitespace and clear section boundaries.
- Thin neutral separators, restrained corner radii, and little or no shadow.
- Compact navigation and simple controls.
- Pill-shaped primary actions where appropriate.
- Minimal motion used only to clarify interaction.

The implementation must not use Apple logos, product imagery, copied marketing text, or other Apple brand assets.

## Guiding Rules

1. Content comes first. Backgrounds remain calm and do not compete with data or actions.
2. Avoid decorative gradients, glowing shapes, visible grid textures, heavy blur, and exaggerated glass effects.
3. Avoid turning every region into a floating card. Use spacing and separators before containers.
4. Use one primary accent color consistently.
5. Preserve information density on administrative screens while improving hierarchy and readability.
6. Preserve existing responsive behavior and make touch targets usable on mobile.
7. Keep light and dark themes structurally identical so switching themes changes appearance, not layout.

## Visual Foundation

### Typography

Use the system stack beginning with `-apple-system` and `BlinkMacSystemFont`, with the project's existing sans-serif fonts as fallbacks. Remove conflicting font declarations that prevent the system stack from taking effect.

- Page titles: semibold, tight tracking, responsive sizing.
- Section titles: semibold with less contrast than page titles.
- Body text: regular or medium weight with comfortable line height.
- Metadata: smaller neutral text, never excessively faint.
- Labels: sentence case. Avoid decorative all-caps except where an existing semantic eyebrow benefits from it.

### Light Theme

- Primary background: white.
- Secondary section background: Apple-like soft gray near `#f5f5f7`.
- Primary text: near `#1d1d1f`.
- Secondary text: near `#6e6e73`.
- Accent: Apple-like blue near `#0071e3`.
- Separators: low-contrast black with transparency.

### Dark Theme

- Primary background: black or near-black.
- Secondary section background: graphite.
- Primary text: near-white.
- Secondary text: medium neutral gray with accessible contrast.
- Accent: a brighter blue suited to dark backgrounds.
- Separators: low-contrast white with transparency.

Colors remain semantic CSS variables so components do not hard-code theme-specific values.

### Shape and Depth

- Controls use compact, consistent radii.
- Primary buttons may use full pill radii.
- Cards and dialogs use moderate radii rather than oversized bubbles.
- Shadows are reserved for floating layers such as menus, sheets, and dialogs.
- Static content containers rely on background contrast and separators instead of elevation.
- Sticky navigation may use subtle translucency for readability, but ordinary content remains opaque.

## Architecture

### Shared Design Tokens

`frontend/src/app/globals.css` remains the semantic theme source. Existing color variables are updated for the Apple.com-inspired palette, and shared radius and focus behavior are normalized there.

The route-scoped students prototype and its dedicated route-detection helper are removed. The approved design applies through the shared system instead of a high-specificity override layer.

### Shared UI Primitives

Restyle existing primitives at their source so all consumers remain consistent:

- Buttons and button variants
- Cards and card sections
- Inputs, textareas, selects, comboboxes, and number fields
- Dialogs, alert dialogs, sheets, menus, popovers, and tooltips
- Tables, tabs, pagination, badges, alerts, and separators
- Switches, checkboxes, radio controls, and toggles
- Empty states, skeletons, progress, and loading indicators

Component APIs and behavior remain stable unless a minor type-safe styling prop is required by an existing use case.

### Application Shells

Apply the shared visual system to:

- Public landing page
- Staff authentication pages
- Staff school-selection workspace
- School sidebar and mobile navigation
- Staff content pages and profiles
- Student and guardian portal login and authenticated shells
- Invitation and activation flows
- Legal pages
- Shared footer and toast presentation

Existing layout structure remains intact. Shell-specific edits remove visual treatments that conflict with the shared system and align spacing, widths, and navigation states.

### Route-Level Cleanup

Route components receive targeted styling cleanup only when they contain hard-coded visual decisions that bypass shared primitives. Business logic, data fetching, permissions, server actions, state management, validation, and navigation remain unchanged.

No backend files are modified.

## Interaction and Product States

The restyle must preserve and clearly present:

- Loading and skeleton states
- Empty and no-result states
- Error and permission-denied states
- Disabled and pending controls
- Form validation and server errors
- Destructive-action confirmation
- Optimistic and refreshed data states
- Pagination and search state
- Mobile navigation and dialogs

Primary actions receive the strongest emphasis. Secondary and destructive actions remain distinguishable without introducing unnecessary color.

## Accessibility and Responsiveness

- Maintain visible keyboard focus indicators.
- Preserve semantic markup and accessible names.
- Meet practical contrast requirements in both themes.
- Do not rely on color alone for state.
- Honor reduced-motion preferences.
- Preserve readable layouts from small mobile screens through wide desktops.
- Keep interactive targets comfortable on touch devices.
- Continue using US English formatting wherever English locale-specific formatting is required.

## Verification

### Automated

- Existing unit tests
- New focused regression tests only where visual-system logic introduces meaningful behavior
- Full ESLint run
- TypeScript check
- Next.js production build
- `git diff --check`

### Visual and Functional

Review representative routes in light and dark modes at mobile and desktop widths:

- Landing page
- Staff login or registration
- School selection
- School dashboard
- Students list and student profile
- Staff, roles, settings, and logs
- Course or assignment detail
- Portal login and authenticated portal
- Legal document

For each representative route, verify navigation, forms, dialogs, loading, empty, error, destructive, and permission-sensitive states when available.

## Delivery Strategy

Implement from the center outward:

1. Remove the route-specific students experiment.
2. Establish global tokens and typography.
3. Restyle shared primitives.
4. Restyle application shells.
5. Clean up conflicting route-level classes.
6. Verify representative routes and all automated checks.

This order makes shared changes visible early, reduces duplicated work, and keeps route-level edits focused.

## Non-Goals

- No route, workflow, or information-architecture redesign.
- No backend or API changes.
- No new product features.
- No dependency additions unless an unexpected hard requirement is discovered and approved.
- No copying of Apple content or assets.
- No conversion into an iOS or macOS interface.
