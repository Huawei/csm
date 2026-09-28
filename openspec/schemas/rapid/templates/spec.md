<!--
Delta spec template for changes.

This template demonstrates four types of delta sections. Use them as needed:
- ADDED / MODIFIED / REMOVED / RENAMED

Strict formatting rules (enforced by OpenSpec validation):
- Every Requirement MUST contain the word `SHALL` or `MUST`.
- Each Requirement MUST include at least one `#### Scenario:` section.
- Scenario sections MUST use level-4 headings (`####`). Level-3 headings or bullet points will cause silent failures.
-->

## ADDED Requirements

<!-- New behaviors. List new requirements to be introduced to the capability in this change. -->

### Requirement: <!-- requirement name -->
<!-- requirement text — must include SHALL or MUST -->

#### Scenario: <!-- scenario name -->
- **WHEN** <!-- condition -->
- **THEN** <!-- expected outcome -->

---

## MODIFIED Requirements

<!--
For updates to existing Requirements.
**MUST use the exact normalized header** (case-sensitive comparison after trimming) from openspec/specs/<capability>/spec.md.
Otherwise, delta application during archiving will fail as the target requirement cannot be located.

**MUST paste the full updated content** (do not only list differences).
OpenSpec archive applies MODIFIED entries via full content replacement.
-->

### Requirement: <!-- identical header as in the original spec -->
<!-- full updated requirement text — must include SHALL or MUST -->

#### Scenario: <!-- scenario name (new or revised) -->
- **WHEN** <!-- condition -->
- **THEN** <!-- expected outcome -->

---

## REMOVED Requirements

<!--
For removing existing Requirements.
A Reason and Migration guide MUST be included to help reviewers understand the removal rationale and guide dependent consumers through migration.
-->

### Requirement: <!-- exact header of the requirement to be removed, consistent with the original spec -->

**Reason**: <!-- rationale for removal -->

**Migration**: <!-- adjustments for existing callers and dependencies -->

---

## RENAMED Requirements

<!--
For renaming Requirement headers. Fixed format: use code-fenced headers for FROM / TO.

If a header is renamed alongside content changes:
List the name change under RENAMED, and provide the full updated content under MODIFIED using the new header.

Execution order during archive application: RENAMED → REMOVED → MODIFIED → ADDED
-->

- FROM: `### Requirement: <Old Name>`
- TO: `### Requirement: <New Name>`