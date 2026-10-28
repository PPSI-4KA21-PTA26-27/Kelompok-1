---
name: Editorial Streetwear
colors:
  surface: '#fbf9f5'
  surface-dim: '#dbdad6'
  surface-bright: '#fbf9f5'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#f5f3ef'
  surface-container: '#efeeea'
  surface-container-high: '#eae8e4'
  surface-container-highest: '#e4e2de'
  on-surface: '#1b1c1a'
  on-surface-variant: '#464742'
  inverse-surface: '#30312e'
  inverse-on-surface: '#f2f0ed'
  outline: '#767872'
  outline-variant: '#c7c7c0'
  surface-tint: '#5f5e5d'
  primary: '#000000'
  on-primary: '#ffffff'
  primary-container: '#1c1c1a'
  on-primary-container: '#858382'
  inverse-primary: '#c9c6c4'
  secondary: '#536600'
  on-secondary: '#ffffff'
  secondary-container: '#c6f000'
  on-secondary-container: '#576a00'
  tertiary: '#000000'
  on-tertiary: '#ffffff'
  tertiary-container: '#1d1c16'
  on-tertiary-container: '#87837c'
  error: '#ba1a1a'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#e5e2e0'
  primary-fixed-dim: '#c9c6c4'
  on-primary-fixed: '#1c1c1a'
  on-primary-fixed-variant: '#474745'
  secondary-fixed: '#c9f300'
  secondary-fixed-dim: '#b0d500'
  on-secondary-fixed: '#171e00'
  on-secondary-fixed-variant: '#3e4c00'
  tertiary-fixed: '#e7e2d9'
  tertiary-fixed-dim: '#cbc6be'
  on-tertiary-fixed: '#1d1c16'
  on-tertiary-fixed-variant: '#494740'
  background: '#fbf9f5'
  on-background: '#1b1c1a'
  surface-variant: '#e4e2de'
typography:
  display-lg:
    fontFamily: Bricolage Grotesque
    fontSize: 48px
    fontWeight: '800'
    lineHeight: 52px
    letterSpacing: -0.03em
  display-lg-mobile:
    fontFamily: Bricolage Grotesque
    fontSize: 34px
    fontWeight: '800'
    lineHeight: 38px
    letterSpacing: -0.025em
  headline-lg:
    fontFamily: Bricolage Grotesque
    fontSize: 32px
    fontWeight: '700'
    lineHeight: 36px
    letterSpacing: -0.02em
  headline-lg-mobile:
    fontFamily: Bricolage Grotesque
    fontSize: 26px
    fontWeight: '700'
    lineHeight: 30px
    letterSpacing: -0.02em
  headline-md:
    fontFamily: Bricolage Grotesque
    fontSize: 22px
    fontWeight: '700'
    lineHeight: 26px
    letterSpacing: -0.015em
  price-display:
    fontFamily: Bricolage Grotesque
    fontSize: 20px
    fontWeight: '700'
    lineHeight: 24px
    letterSpacing: -0.01em
  body-lg:
    fontFamily: Geist
    fontSize: 16px
    fontWeight: '400'
    lineHeight: 24px
    letterSpacing: -0.01em
  body-md:
    fontFamily: Geist
    fontSize: 14px
    fontWeight: '400'
    lineHeight: 20px
    letterSpacing: 0em
  body-sm:
    fontFamily: Geist
    fontSize: 12px
    fontWeight: '400'
    lineHeight: 16px
    letterSpacing: 0em
  label-lg:
    fontFamily: Geist
    fontSize: 14px
    fontWeight: '600'
    lineHeight: 18px
    letterSpacing: 0.01em
  label-md:
    fontFamily: Geist
    fontSize: 12px
    fontWeight: '600'
    lineHeight: 16px
    letterSpacing: 0.02em
  label-caps:
    fontFamily: Geist
    fontSize: 11px
    fontWeight: '700'
    lineHeight: 14px
    letterSpacing: 0.06em
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  gutter: 0.75rem
  gutter-desktop: 1.5rem
  margin: 1rem
  margin-desktop: 3rem
  space-xs: 0.25rem
  space-sm: 0.5rem
  space-md: 0.75rem
  space-lg: 1.25rem
  space-xl: 2rem
---

## Brand & Style

The design system embodies modern streetwear editorial commerce: raw, confident, highly disciplined, and refined. It avoids the noisy clutter typical of standard multi-vendor marketplaces in favor of an art-directed, catalog-first aesthetic reminiscent of limited-run print lookbooks and avant-garde fashion houses. 

Targeted at culturally observant, design-conscious streetwear enthusiasts, the interface evokes focus, high value, and instant clarity. Visually, the system marries warm utilitarian minimalism with high-impact, brutalist typography accents. Layouts prioritize generous breathing room, structured grids, and photographic hierarchy over gimmicks. 

Key product principles:
- **Editorial Directness:** Imagery and typography carry the visual weight. UI controls remain secondary until called upon.
- **Intentional Restraint:** Zero extraneous ecommerce bloat. Strict exclusion of wishlists, customer review sections, complex coupon engines, and speculative tracking widgets. The user journey centers exclusively on discovery, sizing, and direct acquisition.
- **Mobile-First Tactility:** Designed natively for one-handed mobile navigation with 44px+ minimum touch targets, deliberate micro-interactions, and full awareness of iOS safe-area insets.

## Colors

The palette establishes an organic, warm editorial backdrop punctuated by an unapologetic, high-visibility accent.

- **Canvas & Surfaces:**
  - Base Background: Warm off-white (`#FBF9F5`) to prevent optical glare and evoke archival paper.
  - Surface Raised / Card Backplate: Tonal warm stone (`#F5F2EB`).
  - Subtle Structural Borders: Pale warm outline (`#E6E2D8`), strictly 1px solid.
- **Typography & Structural Elements:**
  - Ink Primary: Deep charcoal-black (`#141413`) for editorial headings, body text, and solid high-contrast buttons.
  - Ink Secondary / Elevated Dark: Deep charcoal tint (`#1F1F1E`) for nested containers and inverted badges.
  - Ink Muted: Atmospheric stone gray (`#6E6B64`) for secondary metadata, product SKUs, and tabular sizing labels.
- **The Single Accent (Electric Lime):**
  - Accent Pop: High-voltage lime (`#D4FF14` / `#CCFF00`).
  - Strict Rule of Use: Reserved exclusively for primary conversion triggers (e.g., "Add to Bag", checkout actions), active filter states, and low-inventory scarcity badges. It must never be applied indiscriminately or across large layout fills.
- **Feedback & States:**
  - Error: Stark crimson (`#D92D20`) on soft rose (`#FEE4E2`).
  - Success / In Stock: Deep forest charcoal (`#1A3826`) paired with electric lime indicator ticks.

## Typography

The typographic hierarchy pairs the distinct, structural character of **Bricolage Grotesque** with the clinical precision of **Geist**.

- **Editorial Headlines & Numbers:** Bricolage Grotesque commands attention for titles, section drops, hero statements, and all financial figures. All prices rendered in Indonesian Rupiah (e.g., `Rp 489.000`) must utilize tabular figure settings (`font-variant-numeric: tabular-nums`) to ensure strict vertical visual alignment across checkout rows and product grids.
- **Interface & Narrative Body:** Geist delivers neutral legibility across product composition details, shipping disclaimers, sizing matrices, and standard control elements.
- **Caps & Micro-Labels:** `label-caps` is always set in uppercase with `0.06em` tracking to introduce an industrial, garment-tag feel for metadata such as fabric weight, sizing abbreviations, and stock statuses.

## Layout & Spacing

The layout is built upon an adaptive fluid grid prioritizing mobile density and editorial proportions on wide viewports.

- **Mobile Rhythm (Default, < 768px):**
  - Single or 2-column asymmetric grid for product discovery.
  - Page margin is locked to `1rem` (16px), with `0.75rem` (12px) gutters between product tiles.
  - Safe area padding (`env(safe-area-inset-top)` and `env(safe-area-inset-bottom)`) is strictly applied to fixed navigation and bottom utility drawers.
- **Desktop Grid (>= 1024px):**
  - Transitions to an art-directed 4-column product grid with max-width capped at `1440px`.
  - Margins expand to `3rem` (48px) with `1.5rem` (24px) gutters, generating comfortable editorial whitespace.
- **Component Densities:**
  - Vertical stacking uses `space-lg` (20px) between related data groups and `space-xl` (32px) between unrelated sections.
  - Tight internal groupings (such as price next to a stock badge) leverage `space-xs` (4px) or `space-sm` (8px).

## Elevation & Depth

Visual depth is conveyed through tactile surface layering, crisp 1px borders, and frosted translucent backdrops rather than muddy drop shadows.

- **Surface Layering:**
  - Elevation level 0: Base background `#FBF9F5`.
  - Elevation level 1: Cards and elevated modules `#F5F2EB` framed by a subtle 1px border of `#E6E2D8`.
  - Elevation level 2: Modals, size-selection bottom sheets, and slide-over bags `#FFFFFF` with a hairline `#E6E2D8` border.
- **Sticky Header Glassmorphism:**
  - Fixed mobile top bars and tab bars use an ultra-frosted background: `rgba(251, 249, 245, 0.85)` coupled with `backdrop-filter: blur(16px)` and a bottom border of `1px solid rgba(230, 226, 216, 0.7)`.
- **Shadow Philosophy:**
  - Standard cards and list items contain zero drop shadow (`box-shadow: none`).
  - Floating bottom checkout actions and elevated drawers utilize an ultra-diffused atmospheric lift: `0 8px 30px rgba(20, 20, 19, 0.06)`.

## Shapes

The geometric framework establishes clear distinctions between containing surfaces, interactive targets, and metadata badges:

- **Product & Content Cards:** Fixed at `16px` (`rounded-2xl`). This softens heavy photography while maintaining a contemporary architectural silhouette.
- **Standard UI Controls & Inputs:** Inputs, dropdowns, and button containers use `12px` (`0.75rem`), balancing angular structure with thumb-friendly comfort.
- **Tags, Filter Chips & Badges:** Strictly full pill radius (`9999px`). The pill shape signals instant tappability for categorical tags and separates metadata visually from angular product frames.

## Components

### Buttons
- **Primary CTA:** High-voltage electric lime background (`#D4FF14`), `#141413` text, semi-bold 14px Geist, `rounded-xl` (12px), min-height 48px to exceed the 44px touch target requirement. Hover/active state: subtle scale-down (`scale(0.98)`) and high-contrast charcoal border.
- **Secondary Action:** Transparent background, 1px solid `#141413`, text `#141413`.
- **Tertiary / Utility:** Warm stone surface (`#F5F2EB`), 1px solid `#E6E2D8`, text `#141413`.

### Filter & Category Chips
- Pill-shaped (`rounded-full`), height 36px, horizontal padding 16px.
- Unselected: Surface `#F5F2EB`, border 1px solid `#E6E2D8`, text `#6E6B64`.
- Selected: Background `#141413`, border 1px solid `#141413`, text `#D4FF14`.

### Input Fields & Selectors
- Minimum height 48px, background `#FFFFFF`, border 1px solid `#E6E2D8`, radius 12px.
- Typography: 16px font size on mobile (prevents iOS auto-zoom).
- Focus state: Border transitions to `#141413` with no synthetic fuzzy ring.

### Product Cards
- Background `#F5F2EB`, border 1px solid `#E6E2D8`, radius 16px, overflow hidden.
- Media aspect ratio strictly 4:5 (portrait fashion standard).
- Contains image preview, garment name in Geist semi-bold, and Rupiah price displayed in Bricolage Grotesque tabular format.
- Bottom-edge sizing pills dynamically display availability on tap.

### 4-State UI Archetype
Every core dynamic view (Catalog, Cart Drawer, Product Detail) adheres strictly to the following 4 states:
1. **Loading (Skeleton):** Pulsing blocks utilizing base `#F5F2EB` shifting softly to `#E6E2D8` with no harsh gradients.
2. **Empty:** High-comfort blank state on `#FBF9F5` featuring minimalist editorial copywriting (e.g., "YOUR BAG IS EMPTY / DROP 01 ARCHIVE STILL AVAILABLE"), followed by a direct primary return button.
3. **Error with Retry:** Clean outline panel in `#FEE4E2` featuring a concise explanation, accompanied by a high-contrast `#141413` "RETRY REQUEST" action.
4. **Populated:** Full editorial layouts with progressive image loading and instant interactivity.

### Strictly Excluded Elements
In accordance with core requirements:
- No wishlist icons or floating heart overlays.
- No social proof ratings, stars, or written customer review accordions.
- No coupon codes, promo banner carousels, or discount input modals.
- No live map courier tracking or simulated fulfillment animations.