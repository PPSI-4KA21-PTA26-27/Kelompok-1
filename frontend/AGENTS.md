
# AGENTS.md — Frontend Engineering & Design Rules

## 1. Mission

You are the senior frontend engineer for Hii Stranger.

Your task is to implement and maintain the frontend inside:

`frontend/`

The original UI design references are located at:

`C:\Users\Lananuranf\Lanjutan\Komersial\hii-stranger1\ui-hii-stranger`

This reference directory contains the user's original HTML and PNG design files.

IMPORTANT:
- `frontend/` is the application implementation directory.
- `ui-hii-stranger/` is the design reference directory.
- Never confuse these two directories.
- Never replace the original design with an invented design.

## 2. Mandatory Design Inspection

Before implementing or changing any UI:

1. Inspect the repository root.
2. Inspect `frontend/` and identify its framework and structure.
3. Inspect `ui-hii-stranger/` and discover all relevant design files.
4. Read the supplied HTML files.
5. Open and inspect the supplied PNG images.
6. Identify which references correspond to each page or component.
7. Inspect related CSS, JavaScript, fonts, icons, and image assets.
8. Compare the references with the existing frontend implementation.
9. Identify the minimum changes needed to reproduce the design.

Do not start coding before inspecting the relevant references.

Do not assume that a particular HTML file or PNG is the only
reference. Discover the actual files first.

## 3. Design Source of Truth

The user's supplied design is the primary visual reference.

Use the references in this order:

1. Explicit requirements from the user.
2. The relevant original PNG design.
3. The corresponding HTML and CSS reference implementation.
4. Existing application architecture and functionality.
5. Established framework conventions.

Use the PNG to verify visual appearance and the HTML to understand
structure, content, styling, and intended interactions.

If references conflict, investigate the conflict instead of
silently inventing a solution.

Never delete, overwrite, or modify the original reference files
unless the user explicitly requests it.

## 4. Strict Anti-Slop Rules

These rules are mandatory.

- Do not invent a new design when a reference already exists.
- Do not redesign the page based on personal preference.
- Do not add unrequested sections, cards, buttons, or features.
- Do not change the original color palette arbitrarily.
- Do not replace the intended typography with generic styling.
- Do not add unnecessary gradients, shadows, animations, or effects.
- Do not replace supplied assets with unrelated images or icons.
- Do not use arbitrary spacing and dimensions when the references
  provide enough information to reproduce them.
- Do not add unnecessary dependencies.
- Do not overengineer simple UI components.
- Do not create duplicate components or styles unnecessarily.
- Do not rewrite unrelated application code.
- Do not use mock data as a substitute for real functionality.
- Do not leave fake buttons, dead links, or placeholder interactions.
- Do not claim completion without verifying the result.

Prefer the smallest complete implementation that faithfully
reproduces the user's intended design.

## 5. Implementation Boundaries

All application implementation work belongs in:

`frontend/`

All original design references must remain in:

`ui-hii-stranger/`

Rules:

- Read reference files without moving them.
- Do not copy the entire reference directory into the application.
- Do not treat standalone reference HTML as the application itself.
- Translate the design into the existing framework and component system.
- Reuse relevant assets when their paths and formats are appropriate.
- Keep application code separate from reference material.
- Do not change backend code for a frontend-only task unless necessary
  and explicitly justified.

The reference path is a Windows path. Resolve it relative to the
repository root when possible so the project remains portable.

Expected repository structure:

`hii-stranger1/`
- `backend/`
- `frontend/`
- `rules-direction/`
- `ui-hii-stranger/`

Verify the actual structure before editing.

## 6. Framework and Code Quality

Inspect `frontend/package.json`, the lockfile, and source files
before choosing implementation techniques.

If the project uses Next.js:

- Follow the installed Next.js and React versions.
- Preserve the existing App Router or Pages Router architecture.
- Prefer Server Components when client-side interactivity is unnecessary.
- Use Client Components only where interaction requires them.
- Keep client boundaries small.
- Preserve existing routes and navigation.
- Follow the existing data-fetching and caching strategy.
- Never expose server secrets to client-side code.

For TypeScript:

- Avoid `any` unless there is a justified reason.
- Prefer accurate, explicit types.
- Do not silence type errors with unsafe casts.
- Validate external and untrusted data appropriately.

For styling:

- Use the project's existing styling system.
- Avoid introducing a second styling framework.
- Avoid duplicate CSS and conflicting overrides.
- Use semantic HTML and accessible controls.
- Follow existing component naming conventions.

Do not assume the project uses Tailwind, a particular UI library,
or a particular package manager without checking first.

## 7. Visual Fidelity

Reproduce the reference as closely as practical.

Verify:

- Page layout and container width.
- Element positions and alignment.
- Margins, padding, and spacing.
- Font family, size, weight, and line height.
- Colors and background surfaces.
- Borders, radii, and shadows.
- Image sizes, aspect ratios, and cropping.
- Buttons, inputs, icons, and navigation.
- Content hierarchy and text wrapping.
- Modal and overlay positioning.
- Responsive behavior.

Do not approximate prominent elements when the reference
contains sufficient information to reproduce them.

If the original design contains an unusual layout, preserve it
rather than replacing it with a conventional template.

## 8. Functionality

The frontend must work, not merely look correct.

- Preserve existing functionality.
- Implement interactions according to the reference and requirements.
- Connect controls to real behavior where applicable.
- Validate form input.
- Handle loading, success, and error states appropriately.
- Do not invent API endpoints.
- Inspect the backend API before integrating frontend requests.
- Never hardcode fake success responses for unfinished features.
- Do not remove working behavior to simplify the UI.

If the intended behavior is unclear, inspect existing code first.
Ask for clarification when a consequential decision cannot be
resolved from the available evidence.

## 9. Responsive Design and Accessibility

- Follow supplied responsive references when available.
- Otherwise, adapt the original design sensibly to smaller screens.
- Prevent horizontal overflow and overlapping elements.
- Preserve the visual hierarchy across viewport sizes.
- Use semantic HTML.
- Support keyboard navigation and visible focus states.
- Provide accessible names for controls.
- Associate form labels with inputs.
- Use suitable alternative text for meaningful images.
- Maintain sufficient contrast.

Do not redesign the desktop version merely to solve a mobile issue.

## 10. Verification Workflow

After implementing a page:

1. Run the frontend using its existing development command.
2. Open the page in a browser when browser access is available.
3. Compare the rendered page against the original PNG.
4. Use matching viewport dimensions for comparison.
5. Check spacing, alignment, typography, colors, and assets.
6. Test interactions and responsive behavior.
7. Check browser console errors and failed network requests.
8. Run the project's available lint, type-check, test, and build commands.
9. Fix regressions introduced by the implementation.

Do not claim that a visual comparison was completed unless
the rendered page was actually inspected.

If browser inspection is unavailable, state that limitation.

## 11. Minimal Diff and Safety

- Modify only files required by the task.
- Preserve unrelated user changes.
- Review the current Git diff before editing.
- Do not delete original HTML or PNG files.
- Do not rewrite configuration without a valid reason.
- Do not upgrade dependencies unnecessarily.
- Do not remove existing pages or components without approval.
- Do not run destructive commands that could discard user work.

## 12. Definition of Done

A task is complete only when:

- The requested UI has been implemented in `frontend/`.
- The relevant original references in `ui-hii-stranger/` were inspected.
- The design matches the references as closely as practical.
- Existing functionality remains intact.
- Relevant interactions work correctly.
- Responsive behavior has been considered.
- Applicable checks have actually been run.
- Remaining visual differences and limitations are documented.

## 13. Final Report

At the end of each task, report:

1. What was implemented.
2. Which HTML and PNG references were used.
3. Which files inside `frontend/` were changed.
4. Whether visual comparison was performed.
5. Which checks were executed and their actual results.
6. Remaining differences, failures, or limitations.

Be concise, factual, and honest.

Never claim that code, tests, or visual output were verified
when they were not.
