## Context

<!-- Background and current state -->

## Success Criteria

<!-- How to measure if the problem is solved -->


## Goals / Non-Goals

**Goals:**
<!-- What this design aims to achieve -->

**Non-Goals:**
<!-- What is explicitly out of scope -->


## Integration Points

| Existing Module | Integration Method | Required Changes |
|-----------------|--------------------|------------------|
|                 |                    |                  |

---

## Class Diagrams

<!-- Use subheadings for different class diagrams if needed, e.g., ### Key Classes, ### Factory Pattern -->

### Key Classes

```mermaid
classDiagram
    class User {
        -id: string
        -name: string
    }
    class ExportJob {
        -id: string
        -format: string
        -status: string
    }
    class DataRecord {
        -id: string
        -content: string
    }
    User "1" --> "*" ExportJob: initiates
    ExportJob --> "*" DataRecord: exports
```

## Sequence Diagrams

<!-- Use subheadings for different business scenarios, e.g., ### User Registration, ### Order Processing. Participant names should be sourced from AGENTS.md at project root. -->

### CSV Export

```mermaid
sequenceDiagram
    actor User
    participant ES
    participant FG
    User ->> ES: Request Export
    ES ->> FG: Generate CSV
    FG -->> ES: CSV File
    ES -->> User: Download
```

## Activity Diagrams

<!-- Use subheadings for different business scenarios, e.g., ### CSV Export Flow, ### Batch Import Flow -->

### CSV Export Flow

```mermaid
graph TB
    A[User Request] --> B[Validate Request]
    B --> C{Valid?}
    C -->|Yes| D[Generate CSV]
    C -->|No| E[Return Error]
    D --> F[Store File]
    F --> G[Return Download Link]
    E --> H[End]
    G --> H
```

## Decisions

<!-- Key design decisions and rationale -->

## Risks / Trade-offs

<!-- Known risks and trade-offs -->

## Migration Plan

<!-- Steps to deploy, rollback strategy -->

## Open Questions

| Question | Priority        | Status  |
|----------|-----------------|---------|
|          | High/Medium/Low | Pending |
