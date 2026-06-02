# UI Consistency Design

**Date**: 2026-06-02
**Status**: approved

## Problem

The frontend has 12 categories of visual inconsistency across pages and components:
- Accent colors split across amber, blue, and indigo
- Buttons use 5+ different style conventions for the same intent
- Border radius tiers applied unpredictably (lg/xl/2xl)
- Icon library pages lack the warm background present on other pages
- Page max-width set differently on every page (6 different values)
- Focus rings, modals, loading spinners, and error notifications diverge

## Design Tokens

| Token | Value |
|---|---|
| Primary color | `amber-500` (#f59e0b) |
| Primary light | `amber-50`, `amber-100`, `amber-200` |
| Page bg — workspace/landing | `#FFFDF7` |
| Page bg — icon library | `#F9FAFB` |
| Card bg | `white` |
| Card border | `border border-gray-200` |
| Text primary | `text-gray-900` |
| Text body | `text-gray-600` |
| Text secondary | `text-gray-500` |
| Text inactive | `text-gray-400` |
| Container radius (large) | `rounded-2xl` — page main cards, modals, landing cards |
| Card radius | `rounded-xl` — sub-cards, panels, SvgCanvas, IconCard |
| Button / input radius | `rounded-lg` |
| Focus ring | `focus:ring-2 focus:ring-amber-500/20` |
| Page width — narrow | `max-w-3xl` (Convert, AI Generate) |
| Page width — wide | `max-w-6xl` (Editor, Icon Library, Library, Preview) |
| Page title | `text-xl font-bold text-gray-900` (sub-pages); `text-4xl font-extrabold` (landing only) |
| Grid gap | `gap-4` |

## Component Standards

### Buttons

**Primary large** (CTA, hero, generate):
`rounded-lg bg-amber-500 text-gray-900 font-bold shadow-md shadow-amber-200 hover:-translate-y-0.5`

**Primary small** (toolbar, card actions):
`rounded-lg bg-amber-500 text-white font-medium hover:bg-amber-600`

**Secondary outline**:
`rounded-lg border border-gray-300 text-gray-700 hover:bg-gray-50`

### Modals

Overlay: `bg-black/40`
Dialog: `rounded-2xl bg-white p-6`

### Loading

Shared `LoadingSpinner` component: change `border-t-indigo-600` to `border-t-amber-500`.
All pages use this shared spinner. Remove custom inline spinner from AiGeneratePage.

### Feedback

**Toast (operation result)**: `fixed top-20 right-4 z-50 px-4 py-2 bg-gray-800 text-white text-sm rounded-lg shadow-lg` (success); `bg-red-500` (error)

**Inline error box** (form validation): `rounded-lg bg-red-50 border border-red-200 p-4 text-sm text-red-700`

### Tags

`px-1.5 py-0.5 rounded text-[10px] bg-amber-50 text-amber-600`

### Tabs

Container: `rounded-xl bg-gray-100 p-1`
Active: `rounded-lg bg-white text-gray-900 shadow-sm`
Inactive: `text-gray-500 hover:text-gray-700`

### Inputs

`rounded-lg border border-gray-200 px-3 py-2 text-sm focus:ring-2 focus:ring-amber-500/20 focus:outline-none`

### Textarea

Same as input + `resize-none`

### Nav links

Default: `text-gray-600 hover:bg-gray-100`
Active/current: `text-amber-600 bg-amber-50`

### Cards

**Feature card** (LandingPage): `rounded-2xl border border-gray-200 bg-white p-5`

**Content card** (IconCard, ConversionCard): `rounded-xl border border-gray-200 bg-white hover:shadow-md`

**Panel** (SidePanel, SvgCanvas): `rounded-xl border border-gray-200 bg-white p-4`

## Layout Standards

| Rule | Spec |
|---|---|
| Workspace / landing bg | `bg-[#FFFDF7]` |
| Icon library bg | `bg-[#F9FAFB]` |
| Page padding | `px-6 py-8` (provided by WorkspaceShell or page wrapper) |
| Grid gap | `gap-4` everywhere |

Icon library pages (`/icons`, `/icons/:id`) must share the same page wrapper (Nav + bg + padding) as workspace pages. Option: extract a `PageShell` component or include them in workspace route structure.

## Color Migration

Remove all `blue-*` and `indigo-*` accent usages and replace with amber equivalents:

| Before | After |
|---|---|
| `bg-blue-600`, `bg-indigo-600` (buttons) | `bg-amber-500` |
| `text-blue-600`, `text-indigo-600` (links, accents) | `text-amber-600` |
| `bg-blue-100 text-blue-700` (tag chips) | `bg-amber-50 text-amber-600` |
| `focus:ring-blue-500/20` | `focus:ring-amber-500/20` |
| `border-t-indigo-600` (spinner) | `border-t-amber-500` |
| `border-indigo-500 bg-indigo-50` (drop zone) | `border-amber-500 bg-amber-50` |
| `hover:border-amber-200` (card hover) | remove, use `hover:shadow-md` only |

## Files Changed

- `tailwind.config.js` — add custom tokens
- `src/components/LoadingSpinner.tsx` — amber spinner
- `src/components/Navbar.tsx` — nav link active state
- `src/components/PublishDialog.tsx` — modal overlay, radius, blue→amber, input focus
- `src/components/EmailLoginModal.tsx` — input focus
- `src/components/UsageLimitModal.tsx` — input focus if any
- `src/components/WorkspaceShell.tsx` — verify bg/padding
- `src/components/SearchBar.tsx` — blue→amber, tag chips
- `src/components/ErrorBoundary.tsx` — indigo→amber button
- `src/pages/LandingPage.tsx` — verify styles match tokens
- `src/pages/AiGeneratePage.tsx` — title downgrade, remove custom spinner, remove px-4 override
- `src/pages/ConvertPage.tsx` — inline error box
- `src/pages/PreviewPage.tsx` — indigo→amber buttons
- `src/pages/IconLibraryPage.tsx` — bg, blue→amber tabs/focus
- `src/pages/IconDetailPage.tsx` — bg, blue→amber links
- `src/pages/LibraryPage.tsx` — blue→amber tabs
- `src/pages/EditorPage.tsx` — verify styles
- `src/features/svg-editor/components/ThemeReplacer.tsx` — indigo→amber
- `src/features/svg-editor/components/DropZone.tsx` — indigo→amber
