# UI Consistency Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Unify visual design language across all pages by replacing scattered blue/indigo accents with amber, standardizing button/container/input styles to shared design tokens, and applying consistent layout wrappers.

**Architecture:** Class-level CSS-only changes across ~20 frontend files. No logic changes. No new dependencies. All changes are Tailwind class string substitutions following the design tokens defined in specs/2026-06-02-ui-consistency-design.md.

**Tech Stack:** React 18, TypeScript, Tailwind CSS 3, Vite

---

### Task 1: Design tokens — tailwind.config.js

**Files:**
- Modify: `web/tailwind.config.js`

Add custom color/radius tokens so components can reference named values instead of hardcoded hex.

- [ ] **Step 1: Add custom tokens to tailwind.config.js**

```js
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Nunito', 'system-ui', 'sans-serif'],
      },
      colors: {
        brand: {
          50: '#fffbeb',
          100: '#fef3c7',
          200: '#fde68a',
          500: '#f59e0b',
          600: '#d97706',
          700: '#b45309',
        },
        page: {
          warm: '#FFFDF7',
          cool: '#F9FAFB',
        },
      },
      borderRadius: {
        container: '1rem',   // rounded-2xl equivalent
        card: '0.75rem',     // rounded-xl equivalent
        control: '0.5rem',   // rounded-lg equivalent
      },
    },
  },
  plugins: [],
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `npx tsc -p web/tsconfig.json --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/tailwind.config.js
git commit -m "feat: add custom design tokens to tailwind config"
```

---

### Task 2: LoadingSpinner — indigo → amber

**Files:**
- Modify: `web/src/components/LoadingSpinner.tsx:4`

- [ ] **Step 1: Replace spinner color**

Change line 4 from:
```tsx
<div className="h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-indigo-600" />
```
to:
```tsx
<div className="h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-amber-500" />
```

- [ ] **Step 2: Verify TypeScript compiles**

Run: `npx tsc -p web/tsconfig.json --noEmit`
Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add web/src/components/LoadingSpinner.tsx
git commit -m "fix: change loading spinner accent from indigo to amber"
```

---

### Task 3: ErrorBoundary — indigo → amber button

**Files:**
- Modify: `web/src/components/ErrorBoundary.tsx:32`

- [ ] **Step 1: Replace button color**

Change line 32 from:
```tsx
className="mt-4 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500 transition-colors"
```
to (primary small button spec):
```tsx
className="mt-4 rounded-lg bg-amber-500 px-4 py-2 text-sm font-medium text-white hover:bg-amber-600 transition-colors"
```

- [ ] **Step 2: Commit**

```bash
git add web/src/components/ErrorBoundary.tsx
git commit -m "fix: change ErrorBoundary retry button from indigo to amber"
```

---

### Task 4: DropZone — indigo → amber

**Files:**
- Modify: `web/src/components/DropZone.tsx:45,61`

- [ ] **Step 1: Replace drag state color**

Change line 45 from:
```tsx
${dragging ? 'border-indigo-500 bg-indigo-50' : 'border-gray-300 hover:border-gray-400'}
```
to:
```tsx
${dragging ? 'border-amber-500 bg-amber-50' : 'border-gray-300 hover:border-gray-400'}
```

- [ ] **Step 2: Replace upload text color**

Change line 61 from:
```tsx
<span className="font-semibold text-indigo-600">Click to upload</span> or drag and drop
```
to:
```tsx
<span className="font-semibold text-amber-600">Click to upload</span> or drag and drop
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/DropZone.tsx
git commit -m "fix: change DropZone accent from indigo to amber"
```

---

### Task 5: ThemeReplacer — indigo → amber button

**Files:**
- Modify: `web/src/features/svg-editor/components/ThemeReplacer.tsx:41`

- [ ] **Step 1: Replace button color**

Change line 41 from:
```tsx
className="w-full py-1.5 text-sm font-medium text-white bg-indigo-500 hover:bg-indigo-600 disabled:bg-gray-300 rounded-lg transition-colors"
```
to:
```tsx
className="w-full py-1.5 text-sm font-medium text-white bg-amber-500 hover:bg-amber-600 disabled:bg-gray-300 rounded-lg transition-colors"
```

- [ ] **Step 2: Commit**

```bash
git add web/src/features/svg-editor/components/ThemeReplacer.tsx
git commit -m "fix: change ThemeReplacer button from indigo to amber"
```

---

### Task 6: SearchBar — blue → amber

**Files:**
- Modify: `web/src/components/SearchBar.tsx:38,47-51`

- [ ] **Step 1: Replace focus ring**

Change line 38 from:
```tsx
className="w-full px-4 py-2.5 rounded-lg border border-gray-200 bg-white text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-400"
```
to:
```tsx
className="w-full px-4 py-2.5 rounded-lg border border-gray-200 bg-white text-sm focus:outline-none focus:ring-2 focus:ring-amber-500/20 focus:border-amber-400"
```

- [ ] **Step 2: Replace active tag button color**

Change lines 47-51 from:
```tsx
className={`px-2.5 py-1 rounded-full text-xs font-medium transition-colors ${
  selectedTags.includes(tag.slug)
    ? 'bg-blue-600 text-white'
    : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
}`}
```
to:
```tsx
className={`px-2.5 py-1 rounded-full text-xs font-medium transition-colors ${
  selectedTags.includes(tag.slug)
    ? 'bg-amber-500 text-white'
    : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
}`}
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/SearchBar.tsx
git commit -m "fix: change SearchBar focus ring and tags from blue to amber"
```

---

### Task 7: PublishDialog — blue → amber + modal spec alignment

**Files:**
- Modify: `web/src/components/PublishDialog.tsx`

Changes:
- overlay: `bg-black/30` → `bg-black/40`
- dialog: `rounded-xl` → `rounded-2xl`
- focus rings: `blue-500/20` → `amber-500/20`
- form button: `bg-blue-600 text-white hover:bg-blue-700` → `bg-amber-500 text-gray-900 font-bold shadow-md shadow-amber-200 hover:-translate-y-0.5`
- tag chips: `bg-blue-100 text-blue-700` → `bg-amber-50 text-amber-600`
- tag close button: `text-blue-400 hover:text-blue-600` → `text-amber-400 hover:text-amber-600`

- [ ] **Step 1: Fix overlay + dialog radius**

Line 35, change:
```tsx
<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30" onClick={onClose}>
```
to:
```tsx
<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40" onClick={onClose}>
```

Line 36, change:
```tsx
<div className="bg-white rounded-xl shadow-xl p-6 w-full max-w-md mx-4" onClick={e => e.stopPropagation()}>
```
to:
```tsx
<div className="bg-white rounded-2xl shadow-xl p-6 w-full max-w-md mx-4" onClick={e => e.stopPropagation()}>
```

- [ ] **Step 2: Fix input focus rings (3 occurrences)**

Line 48, change:
```tsx
className="w-full px-3 py-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500/20"
```
to:
```tsx
className="w-full px-3 py-2 rounded-lg border border-gray-200 text-sm focus:outline-none focus:ring-2 focus:ring-amber-500/20"
```

Same change on lines 60 and 83 (the tag input and theme input).

- [ ] **Step 3: Fix submit button**

Line 94, change:
```tsx
<button type="submit" className="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700">发布</button>
```
to (primary small button spec):
```tsx
<button type="submit" className="px-4 py-2 text-sm bg-amber-500 text-white font-medium rounded-lg hover:bg-amber-600">发布</button>
```

- [ ] **Step 4: Fix tag chips**

Line 68, change:
```tsx
<span key={i} className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs bg-blue-100 text-blue-700">
  {t.name}
  <button type="button" onClick={() => removeTag(i)} className="text-blue-400 hover:text-blue-600">&times;</button>
</span>
```
to:
```tsx
<span key={i} className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs bg-amber-50 text-amber-700">
  {t.name}
  <button type="button" onClick={() => removeTag(i)} className="text-amber-400 hover:text-amber-600">&times;</button>
</span>
```

- [ ] **Step 5: Commit**

```bash
git add web/src/components/PublishDialog.tsx
git commit -m "fix: change PublishDialog from blue to amber, align modal spec"
```

---

### Task 8: PreviewPage — indigo → amber button

**Files:**
- Modify: `web/src/pages/PreviewPage.tsx:109`

- [ ] **Step 1: Replace download button color**

Change lines 109 from:
```tsx
className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500 transition-colors"
```
to (primary small button spec):
```tsx
className="inline-flex items-center gap-2 rounded-lg bg-amber-500 px-4 py-2 text-sm font-medium text-white hover:bg-amber-600 transition-colors"
```

- [ ] **Step 2: Commit**

```bash
git add web/src/pages/PreviewPage.tsx
git commit -m "fix: change PreviewPage download button from indigo to amber"
```

---

### Task 9: LibraryPage — blue → amber tabs and links

**Files:**
- Modify: `web/src/pages/LibraryPage.tsx:82,88,135`

- [ ] **Step 1: Replace tab button active color**

Change line 82 from:
```tsx
className={`px-4 py-1.5 text-sm font-medium transition-colors ${tab === 'conversions' ? 'bg-blue-600 text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}`}
```
to:
```tsx
className={`px-4 py-1.5 text-sm font-medium transition-colors ${tab === 'conversions' ? 'bg-amber-500 text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}`}
```

Change line 88 from:
```tsx
className={`px-4 py-1.5 text-sm font-medium transition-colors ${tab === 'saved' ? 'bg-blue-600 text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}`}
```
to:
```tsx
className={`px-4 py-1.5 text-sm font-medium transition-colors ${tab === 'saved' ? 'bg-amber-500 text-white' : 'bg-white text-gray-600 hover:bg-gray-50'}`}
```

- [ ] **Step 2: Replace download link color**

Change line 135 from:
```tsx
className="text-xs text-blue-600 hover:text-blue-800"
```
to:
```tsx
className="text-xs text-amber-600 hover:text-amber-700"
```

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/LibraryPage.tsx
git commit -m "fix: change LibraryPage tabs and links from blue to amber"
```

---

### Task 10: IconDetailPage — blue → amber link

**Files:**
- Modify: `web/src/pages/IconDetailPage.tsx:39,64`

- [ ] **Step 1: Replace back link color**

Change line 39 from:
```tsx
className="text-blue-600 hover:text-blue-800 text-sm"
```
to:
```tsx
className="text-amber-600 hover:text-amber-700 text-sm"
```

- [ ] **Step 2: Replace tag chip style to match design spec**

Change line 64 from:
```tsx
<span key={t.slug} className="inline-block px-2 py-0.5 rounded text-xs bg-gray-100 text-gray-600 mr-1 mb-1">
```
to:
```tsx
<span key={t.slug} className="inline-block px-2 py-0.5 rounded text-xs bg-amber-50 text-amber-600 mr-1 mb-1">
```

- [ ] **Step 3: Commit**

```bash
git add web/src/pages/IconDetailPage.tsx
git commit -m "fix: change IconDetailPage link and tags from blue/gray to amber"
```

---

### Task 11: Navbar — nav link active state

**Files:**
- Modify: `web/src/components/Navbar.tsx` — already uses amber for AI link (line with `text-amber-600`), so only change is to verify the AI generate active nav link follows the spec. No changes needed since Navbar already uses amber.

This task is a no-op — Navbar already conforms.

---

### Task 12: AiGeneratePage — title + spinner + padding

**Files:**
- Modify: `web/src/pages/AiGeneratePage.tsx`

- [ ] **Step 1: Downgrade page title**

Change from:
```tsx
<h1 className="text-2xl font-extrabold text-gray-900 mb-6">AI 生成图标</h1>
```
to (sub-page title spec):
```tsx
<h1 className="text-xl font-bold text-gray-900 mb-6">AI 生成图标</h1>
```

- [ ] **Step 2: Remove custom px-4 override on outer container**

Change from:
```tsx
<div className="mx-auto max-w-3xl px-4 py-8">
```
to (WorkspaceShell already provides px-6):
```tsx
<div className="mx-auto max-w-3xl">
```

- [ ] **Step 3: Replace custom inline spinner with LoadingSpinner**

In the phase 2 (generating) block, replace the custom `<div>` spinner and text with the shared LoadingSpinner component:

Import at top (line 4):
```tsx
import LoadingSpinner from '../components/LoadingSpinner'
```

Replace the generating phase JSX block (currently lines ~117-123) from:
```tsx
{phase === 'generating' && (
  <div className="flex flex-col items-center justify-center py-20">
    <div className="relative mb-6">
      <div className="h-16 w-16 rounded-full border-4 border-amber-100 border-t-amber-500 animate-spin" />
    </div>
    <p className="text-lg font-medium text-gray-700">AI 正在为你设计图标...</p>
    <p className="text-sm text-gray-400 mt-2">这可能需要 10-30 秒</p>
  </div>
)}
```
to:
```tsx
{phase === 'generating' && (
  <LoadingSpinner label="AI 正在为你设计图标..." />
)}
```

- [ ] **Step 4: Verify TypeScript compiles**

Run: `npx tsc -p web/tsconfig.json --noEmit`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add web/src/pages/AiGeneratePage.tsx
git commit -m "fix: unify AiGeneratePage title, spinner, and padding with design spec"
```

---

### Task 13: LandingPage — card hover consistency

**Files:**
- Modify: `web/src/pages/LandingPage.tsx:124`

Current LandingPage workflow cards use `hover:-translate-y-1 hover:shadow-lg hover:border-amber-200`. Per card spec, change to `hover:shadow-md` only.

- [ ] **Step 1: Simplify card hover**

Change line 124 from:
```tsx
className="group flex items-start gap-4 rounded-2xl border border-gray-200 bg-white p-5 transition-all duration-300 hover:-translate-y-1 hover:shadow-lg hover:border-amber-200"
```
to:
```tsx
className="group flex items-start gap-4 rounded-2xl border border-gray-200 bg-white p-5 transition-shadow hover:shadow-md"
```

- [ ] **Step 2: Commit**

```bash
git add web/src/pages/LandingPage.tsx
git commit -m "fix: unify LandingPage card hover with design spec"
```

---

### Task 14: Icon pages — background + padding wrapper

**Files:**
- Modify: `web/src/App.tsx`

IconLibraryPage and IconDetailPage currently render without a shared background or padding wrapper (they get white background by default). Per design, they should have `bg-[#F9FAFB]` and consistent padding.

Wrap both icon routes in a shared layout:

- [ ] **Step 1: Add wrapper in App.tsx**

Change the icon routes (lines ~53-54) from:
```tsx
<Route path="/icons" element={<><Navbar /><IconLibraryPage /><Footer /></>} />
<Route path="/icons/:id" element={<><Navbar /><IconDetailPage /><Footer /></>} />
```
to:
```tsx
<Route path="/icons" element={<><Navbar /><div className="min-h-screen bg-[#F9FAFB]"><div className="mx-auto max-w-6xl px-6 py-8"><IconLibraryPage /></div></div><Footer /></>} />
<Route path="/icons/:id" element={<><Navbar /><div className="min-h-screen bg-[#F9FAFB]"><div className="mx-auto max-w-6xl px-6 py-8"><IconDetailPage /></div></div><Footer /></>} />
```

- [ ] **Step 2: Remove redundant max-w/padding from IconLibraryPage**

In `web/src/pages/IconLibraryPage.tsx:57`, change:
```tsx
<div className="max-w-5xl mx-auto">
```
to:
```tsx
<div>
```
(The wrapper now provides `max-w-6xl`.)

- [ ] **Step 3: Remove redundant max-w from IconDetailPage**

In `web/src/pages/IconDetailPage.tsx:48`, change:
```tsx
<div className="max-w-4xl mx-auto">
```
to:
```tsx
<div>
```
(The wrapper now provides `max-w-6xl`.)

- [ ] **Step 4: Verify TypeScript compiles**

Run: `npx tsc -p web/tsconfig.json --noEmit`
Expected: No errors

- [ ] **Step 5: Commit**

```bash
git add web/src/App.tsx web/src/pages/IconLibraryPage.tsx web/src/pages/IconDetailPage.tsx
git commit -m "fix: add shared background/padding wrapper for icon pages"
```

---

### Task 15: Page width standardization

**Files:**
- Modify: `web/src/pages/PreviewPage.tsx:69` — `max-w-5xl` → `max-w-6xl`
- Modify: `web/src/pages/LibraryPage.tsx:76` — `max-w-5xl` → `max-w-6xl`
- Modify: `web/src/pages/ConvertPage.tsx:78` — `max-w-xl` → `max-w-3xl`

- [ ] **Step 1: Fix PreviewPage max-width**

Change line 69 from:
```tsx
<div className="max-w-5xl mx-auto">
```
to:
```tsx
<div className="max-w-6xl mx-auto">
```

- [ ] **Step 2: Fix LibraryPage max-width**

Change line 76 from:
```tsx
<div className="max-w-5xl mx-auto">
```
to:
```tsx
<div className="max-w-6xl mx-auto">
```

- [ ] **Step 3: Fix ConvertPage max-width**

Change line 78 from:
```tsx
<div className="max-w-xl mx-auto space-y-4">
```
to:
```tsx
<div className="max-w-3xl mx-auto space-y-4">
```

- [ ] **Step 4: Also fix ConvertPage processing state max-width (line 67)**

Change from:
```tsx
<div className="max-w-xl mx-auto">
```
to:
```tsx
<div className="max-w-3xl mx-auto">
```

- [ ] **Step 5: Commit**

```bash
git add web/src/pages/PreviewPage.tsx web/src/pages/LibraryPage.tsx web/src/pages/ConvertPage.tsx
git commit -m "fix: standardize page max-widths to 3xl (narrow) and 6xl (wide)"
```

---

### Task 16: IconLibraryPage — grid gap → 4

**Files:**
- Modify: `web/src/pages/IconLibraryPage.tsx:67`

Current gap is `gap-3`, standard is `gap-4`.

- [ ] **Step 1: Fix gap**

Change line 67 from:
```tsx
<div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
```
to:
```tsx
<div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
```

- [ ] **Step 2: Commit**

```bash
git add web/src/pages/IconLibraryPage.tsx
git commit -m "fix: unify IconLibraryPage grid gap to gap-4"
```

---

### Task 17: IconCard sub-components — verify spec conformance

**Files:**
- Check: `web/src/components/IconCard.tsx`

Read and verify the card uses `rounded-xl` with `hover:shadow-md`. Any deviation from the content card spec (`rounded-xl border border-gray-200 bg-white hover:shadow-md`) should be fixed.

- [ ] **Step 1: Read and verify IconCard styles**

Read the file and check that its classes match `rounded-xl border border-gray-200 bg-white hover:shadow-md`. Fix if needed.

- [ ] **Step 2: Same check for ConversionCard**

Read `web/src/components/ConversionCard.tsx` and verify card styles.

- [ ] **Step 3: Commit any fixes**

If changes were needed:
```bash
git add web/src/components/IconCard.tsx web/src/components/ConversionCard.tsx
git commit -m "fix: verify IconCard and ConversionCard card styles match spec"
```

---

### Task 18: Final TypeScript build + visual smoke list

**Files:** all modified

- [ ] **Step 1: Full TypeScript check**

Run: `npx tsc -p web/tsconfig.json --noEmit`
Expected: No errors

- [ ] **Step 2: Rebuild web container**

Run: `docker compose build web && docker compose up -d web`

- [ ] **Step 3: Visual smoke test checklist**

Open the app at `http://10.100.1.124` and walk through:
- [ ] LandingPage — cards have `hover:shadow-md` only, no translate
- [ ] ConvertPage — drop zone amber accent, max-w-3xl
- [ ] LibraryPage — tabs amber, download link amber
- [ ] PreviewPage — download button amber
- [ ] IconLibraryPage — `bg-[#F9FAFB]`, search focus ring amber, tags amber
- [ ] IconDetailPage — back link amber, tag chips amber-50
- [ ] AiGeneratePage — title xl/bold, spinner shared, no px-4 override
- [ ] EditorPage — theme replacer button amber
- [ ] Loading spinner — amber accent on all pages
- [ ] ErrorBoundary — retry button amber
- [ ] PublishDialog — overlay bg-black/40, rounded-2xl, focus amber, submit amber, tags amber

- [ ] **Step 4: Commit final tweaks**

```bash
git add -A
git commit -m "chore: final UI consistency verification and tweaks"
```
