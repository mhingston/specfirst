# Product Design Protocol

A workflow for product design, from vision through to implementation handoff.

## Stages

| Stage | Output Artifacts |
|-------|-----------------|
| `product-vision` | `product/product-overview.md` |
| `product-roadmap` | `product/product-roadmap.md` |
| `data-model` | `product/data-model/data-model.md` |
| `design-tokens` | `product/design-system/colors.json`, `typography.json` |
| `design-shell` | `product/shell/spec.md`, `src/shell/components/*.tsx` |
| `shape-section` | `product/sections/<id>/spec.md` (repeatable) |
| `sample-data` | `product/sections/<id>/data.json`, `types.ts` (repeatable) |
| `design-screen` | `src/sections/<id>/*.tsx` (repeatable) |
| `screenshot-design` | `product/sections/<id>/*.png` (repeatable) |
| `export-product` | `product-plan/*` |

## Setup

To use this protocol in a new project:

1. Create a new directory and initialize:
   ```bash
   mkdir my-product && cd my-product
   specfirst init
   ```

2. Copy the protocol, templates, and skills:
   ```bash
   cp /path/to/specfirst/examples/product-design/protocol.yaml .specfirst/protocols/
   cp -r /path/to/specfirst/examples/product-design/templates/* .specfirst/templates/
   cp -r /path/to/specfirst/examples/product-design/skills/* .specfirst/skills/
   ```

3. Set the protocol in your config or use the flag:
   ```bash
   specfirst --protocol product-design status
   specfirst --protocol product-design product-vision
   ```

## Section Looping

This protocol supports iteration via `repeatable: true` stages. To design multiple sections:

1. Set `section_id` in your `.specfirst/config.yaml`:
   ```yaml
   custom_vars:
     section_id: invoices
     section_title: Invoices
   ```

2. Run the section stages:
   ```bash
   specfirst shape-section
   # Complete with: specfirst complete shape-section product/sections/invoices/spec.md
   
   specfirst sample-data
   # Complete with: specfirst complete sample-data product/sections/invoices/data.json product/sections/invoices/types.ts
   
   specfirst design-screen
   # Complete with: specfirst complete design-screen src/sections/invoices/InvoiceList.tsx ...
   ```

3. Repeat for each section by changing `section_id`.

## Artifact Structure

```
product/
├── product-overview.md
├── product-roadmap.md
├── data-model/
│   └── data-model.md
├── design-system/
│   ├── colors.json
│   └── typography.json
├── shell/
│   └── spec.md
└── sections/
    ├── invoices/
    │   ├── spec.md
    │   ├── data.json
    │   └── types.ts
    └── dashboard/
        └── ...

src/
├── shell/
│   ├── components/*.tsx
│   └── ShellPreview.tsx
└── sections/
    └── invoices/
        └── *.tsx

product-plan/
└── ... (export output)
```

## Skills

This example uses the `readFile` template helper to include reusable skills:
- `skills/design-principles.md` - Design quality guidelines

Templates reference skills via `{{ readFile "design-principles.md" }}`.

