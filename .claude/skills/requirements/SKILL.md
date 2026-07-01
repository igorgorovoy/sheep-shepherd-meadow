---
name: requirements
description: Requirements Discovery skill for extracting and structuring requirements from raw user input, issues, or vague requests. Use this skill when you need to clarify requirements, extract acceptance criteria, or transform ambiguous requests into structured specifications before planning. Creates requirements documents that feed into classify and planning skills.
---

# Requirements Discovery

Extract, structure, and document requirements from raw user input, GitHub issues, or vague requests. This skill transforms ambiguous requests into clear, actionable requirements documents that planning skills can consume.

## Overview

This skill is used in the **pre-planning phase** of the SDLC to:

- Extract explicit and implicit requirements from raw input
- Identify acceptance criteria and success metrics
- Clarify scope and constraints
- Document assumptions and dependencies
- Ask clarifying questions when requirements are ambiguous
- Produce structured requirements documents

**Position in SDLC workflow:**

```
User Input / Issue → requirements (this skill) → classify → planning skill → implement
```

## Variables

branch_name: Current git branch name (e.g., "feature/new-UI")
descriptive_name: $1 (optional) - Short descriptive name for the requirements doc
input_source: $2 (optional) - "issue", "user", or "conversation" (default: "user")

## Instructions

### Step 1: Gather Raw Input

Collect the raw input from one of these sources:
- **User request**: Direct message from user
- **GitHub issue**: Issue title and body
- **Conversation context**: Previous conversation history

### Step 2: Analyze Input

Examine the input to understand:
- **What is being requested?** (the "what")
- **Why is it needed?** (the business value/motivation)
- **Who will use it?** (the stakeholders/users)
- **What constraints exist?** (technical, time, scope)
- **What is NOT being requested?** (scope boundaries)

### Step 3: Extract Requirements

#### Functional Requirements
Identify what the system should DO:
- User-facing features and behaviors
- System actions and responses
- Data inputs and outputs
- Business rules and logic

#### Non-Functional Requirements
Identify quality attributes:
- Performance expectations
- Security requirements
- Scalability needs
- Usability standards
- Compatibility requirements

#### Acceptance Criteria
Define measurable success criteria:
- Specific, testable conditions
- Expected outcomes
- Edge cases to handle
- Error scenarios

### Step 4: Identify Gaps and Ambiguities

Look for:
- **Missing information**: What's not specified but needed?
- **Ambiguous terms**: What could be interpreted multiple ways?
- **Conflicting requirements**: What contradicts other requirements?
- **Assumptions**: What are we assuming but should verify?

### Step 5: Resolve Ambiguities

When critical information is missing or ambiguous:

**Priority 1: Self-resolve with documented assumptions**
- Make reasonable assumptions based on codebase patterns and domain context
- Document every assumption in the "Assumptions" section of the requirements doc
- Mark assumptions with confidence: `[HIGH]` (safe to proceed) or `[LOW]` (risky)

**Priority 2: Escalate via comms (if `comms_enabled` in state.json)**
Only escalate when ALL of these are true:
- The missing info significantly changes scope or architecture
- No safe assumption can be made from codebase context
- Getting it wrong would require major rework

If comms_enabled: **ask the user directly in Cursor chat**:
"Requirements [{branch_name}]: {specific question}. Options: {option_a}, {option_b}, Decide for me"
Wait for reply. On timeout or "Decide for me": proceed with best assumption, document it.

**Priority 3: Proceed and note open questions**
If comms is not available or times out, proceed with documented assumptions and list open questions in the requirements doc for later review.

### Step 6: Create Requirements Document

Write the requirements document using the `Requirements Format` below.

Save to: `specs/requirements/{descriptive-name}.md`

### Step 7: Create Footprint and Update State

Document the discovery process and update state for workflow continuity.

## Requirements Format

```md
# Requirements: <descriptive name>

## Metadata

branch_name: `{branch-name}`
descriptive_name: `{descriptive-name}`
created: `{ISO 8601 timestamp}`
status: `draft` | `reviewed` | `approved`

## Source

### Input Type
{issue|user|conversation}

### Raw Input
{original text of the request/issue}

### Context
{any additional context gathered}

## Problem Statement

<clearly define the problem or opportunity being addressed>

## Goals and Objectives

### Primary Goal
<the main objective this should achieve>

### Secondary Goals
<additional objectives, if any>

### Success Metrics
<how will we measure success?>

## Stakeholders

### Primary Users
<who will directly use this feature/fix?>

### Secondary Stakeholders
<who else is affected?>

## Requirements

### Functional Requirements

#### FR-1: <requirement title>
- **Description**: <detailed description>
- **Priority**: Must Have | Should Have | Nice to Have
- **Acceptance Criteria**:
  - [ ] <criterion 1>
  - [ ] <criterion 2>

#### FR-2: <requirement title>
...

### Non-Functional Requirements

#### NFR-1: <requirement title>
- **Category**: Performance | Security | Usability | Scalability | Other
- **Description**: <detailed description>
- **Acceptance Criteria**:
  - [ ] <criterion 1>

#### NFR-2: <requirement title>
...

## Scope

### In Scope
<explicitly list what IS included>

### Out of Scope
<explicitly list what is NOT included>

### Future Considerations
<items deferred for later>

## Constraints

### Technical Constraints
<technical limitations or requirements>

### Business Constraints
<time, budget, resource constraints>

### Dependencies
<external dependencies>

## Assumptions

<list assumptions made during requirements gathering>

## Open Questions

<list unresolved questions that need answers>

## Acceptance Criteria Summary

<consolidated list of all acceptance criteria for easy reference>

- [ ] <criterion 1>
- [ ] <criterion 2>
- [ ] <criterion 3>
...

## Recommended Classification

**Suggested Type**: {feature|bug|chore|patch|prototype}
**Reasoning**: <why this classification is recommended>
**Next Skill**: {/design|/feature|/bug|/chore|/patch|prototype-skill}

## Design Assessment

**Needs Design**: {yes|no}
**Complexity**: {low|medium|high}
**Affected Services**: <list services that will be modified>

**Design Triggers** (if needs_design: yes):
- [ ] Affects 3+ services
- [ ] Introduces new architectural patterns
- [ ] Significant database schema changes
- [ ] External integrations required
- [ ] Security implications
- [ ] Performance considerations

**Reasoning**: <why design is/isn't needed>

## Notes

<additional notes, references, or context>
```

## Relevant Files

Focus on understanding the domain before extracting requirements:

- `docs/README.md` - Project overview and context
- `docs/services/README.md` - Services architecture
- `docs/rules/backend/` - Backend conventions and constraints
- `docs/rules/frontend/` - Frontend conventions and constraints

For domain-specific requests, explore relevant service directories to understand existing functionality and constraints.

## Best Practices

### 1. Start Broad, Then Focus

Begin with understanding the overall request before diving into details:
- What is the user trying to achieve?
- What problem are they solving?
- What value does this provide?

### 2. Use Domain Language

Reflect the user's terminology in requirements:
- Understand their vocabulary
- Map technical terms to business concepts
- Keep requirements understandable to non-technical stakeholders

### 3. Be SMART with Acceptance Criteria

Each acceptance criterion should be:
- **S**pecific: Clear and unambiguous
- **M**easurable: Objectively verifiable
- **A**chievable: Technically feasible
- **R**elevant: Connected to the requirement
- **T**estable: Can be validated

### 4. Prioritize Requirements

Use MoSCoW prioritization:
- **Must Have**: Critical for success
- **Should Have**: Important but not critical
- **Could Have**: Nice to have if time permits
- **Won't Have**: Explicitly excluded from scope

### 5. Document Assumptions

Always document assumptions to:
- Enable verification with stakeholders
- Provide context for future decisions
- Reduce misunderstandings

### 6. Keep Scope Manageable

If requirements are too broad:
- Suggest breaking into multiple smaller requests
- Identify MVP (Minimum Viable Product) scope
- Defer non-critical features

## Clarifying Questions Guide

When asking clarifying questions, categorize them:

### Scope Questions
- "Should this include [feature X] or is that separate?"
- "What's the minimum version that would be acceptable?"
- "Are there any features explicitly NOT wanted?"

### User Questions
- "Who is the primary user for this?"
- "What's their technical skill level?"
- "How frequently will they use this?"

### Technical Questions
- "Are there any existing solutions we should integrate with?"
- "Are there performance requirements?"
- "What platforms/browsers need support?"

### Priority Questions
- "What's the deadline or urgency?"
- "Is this blocking other work?"
- "What's the impact if this is delayed?"

## Example: Vague Request to Structured Requirements

### Input
```
"We need better search"
```

### Clarifying Questions
1. "What type of content should be searchable? (users, documents, products, all of the above?)"
2. "What's missing from the current search? (speed, accuracy, filters, relevance?)"
3. "Who uses search most? (end users, admins, both?)"
4. "Are there specific search scenarios that are painful today?"

### After Clarification
```
Title: "Improve product search with filters and relevance ranking"
User Story: "As a customer, I want to filter products by category and price, and see the most relevant results first, so I can find what I need faster"
```

### Extracted Requirements

**FR-1: Category Filtering**
- Users can filter search results by product category
- Multiple categories can be selected
- Acceptance: Filter reduces results to only matching categories

**FR-2: Price Range Filtering**
- Users can specify min/max price range
- Acceptance: Results only show products within range

**FR-3: Relevance Ranking**
- Results sorted by relevance score
- Exact matches ranked higher
- Acceptance: Search for "blue shirt" shows blue shirts before "shirt" or "blue"

**NFR-1: Performance**
- Search results return within 500ms
- Acceptance: 95th percentile response time < 500ms

---

## Footprint and State Management

After creating the requirements document, you MUST create a footprint and update the state file.

### Footprint Creation

Create a footprint file at: `agentic/{branch-name}/footprints/foot-requirements.md`

**Footprint Template:**

```markdown
# Requirements Discovery Footprint

**Date**: {ISO 8601 timestamp}
**Issue/Request**: {original request summary}
**Type**: Requirements Discovery

## Input Analysis

### Input Source
{issue|user|conversation}

### Raw Input
{original text}

### Initial Assessment
{first impressions and key observations}

## Discovery Process

### Step 1: Input Analysis
- **Key Topics Identified**: {list}
- **Initial Classification Guess**: {feature|bug|chore|patch|prototype}
- **Complexity Assessment**: {simple|moderate|complex}

### Step 2: Requirements Extraction
- **Functional Requirements Found**: {count}
- **Non-Functional Requirements Found**: {count}
- **Acceptance Criteria Defined**: {count}

### Step 3: Gap Analysis
- **Missing Information**: {list}
- **Ambiguities Found**: {list}
- **Assumptions Made**: {list}

### Step 4: Clarification (if applicable)
- **Questions Asked**: {list}
- **Answers Received**: {list}
- **Impact on Requirements**: {description}

## Discovery Result

**Requirements Doc**: specs/requirements/{descriptive-name}.md
**Recommended Classification**: {feature|bug|chore|patch|prototype}
**Confidence**: {high|medium|low}

**Key Requirements**:
{list top 3-5 requirements}

**Key Acceptance Criteria**:
{list top 3-5 acceptance criteria}

**Open Questions**:
{list remaining questions}

## Next Steps

**Next Command**: /classify or {specific planning skill}
**Required Context**: specs/requirements/{descriptive-name}.md
**Notes**: {any notes for the next skill}
```

### State File Update

Create or update the state file at: `agentic/{branch-name}/state.json`

```json
{
  "latest_footprint": "agentic/{branch-name}/footprints/foot-requirements.md",
  "latest_command": "requirements",
  "requirements_file_path": "specs/requirements/{descriptive-name}.md",
  "next_command_metadata": {
    "command": "{/design|/classify}",
    "category": "requirements",
    "confidence": "{high|medium|low}",
    "reasoning": "Requirements extracted, ready for {design|classification} and planning",
    "required_context": "specs/requirements/{descriptive-name}.md"
  },
  "next_command": "{/design|/classify}",
  "requirements_timestamp": "{ISO 8601 timestamp}",
  "branch_name": "{branch-name}",
  "descriptive_name": "{descriptive-name}",
  "request_summary": "{brief summary of the request}",
  "recommended_classification": "{feature|bug|chore|patch|prototype}",
  "requirements_status": "draft",
  "needs_design": true|false,
  "complexity": "{low|medium|high}",
  "affected_services": ["{list of services}"]
}
```

---

## Beads Integration

If `bead_id` exists in `agentic/{branch-name}/state.json`, update the bead after requirements are created:

```bash
bd update $bead_id --notes="Requirements created: $requirements_file_path. Complexity: $complexity. Needs design: $needs_design"
```

If `bead_id` is null or missing, skip silently.

---

## Required Actions

After completing requirements discovery, you MUST:

1. **Create Requirements Doc**: Write to `specs/requirements/{descriptive-name}.md`
2. **Create Footprint**: Document process in `agentic/{branch-name}/footprints/foot-requirements.md`
3. **Update State**: Create/update `agentic/{branch-name}/state.json`

## Report

Return a JSON object with the following structure:

```json
{
  "requirements_file_path": "specs/requirements/{descriptive-name}.md",
  "recommended_classification": "{feature|bug|chore|patch|prototype}",
  "needs_design": true|false,
  "complexity": "{low|medium|high}",
  "affected_services": ["{list}"],
  "next_command": "{/design|/classify}",
  "requirements_count": {
    "functional": {count},
    "non_functional": {count},
    "acceptance_criteria": {count}
  },
  "open_questions_count": {count},
  "footprint_path": "agentic/{branch-name}/footprints/foot-requirements.md",
  "state_path": "agentic/{branch-name}/state.json"
}
```

**Field descriptions:**
- `requirements_file_path`: Path to created requirements document
- `recommended_classification`: Suggested classification for routing
- `needs_design`: Whether design phase is recommended (true for complex features)
- `complexity`: Complexity assessment (low/medium/high)
- `affected_services`: List of services that will be modified
- `next_command`: Next skill to invoke (/design if needs_design is true, otherwise /classify)
- `requirements_count`: Count of requirements by type
- `open_questions_count`: Number of unresolved questions
- `footprint_path`: Path to created footprint
- `state_path`: Path to state file

---

## When to Use This Skill

**Use this skill when:**
- User request is vague or ambiguous
- Requirements need clarification before planning
- Scope is unclear or potentially too broad
- Acceptance criteria need to be defined
- You need to document what success looks like
- Multiple interpretations of the request are possible

**Skip this skill when:**
- Requirements are already clear and specific
- Request is a simple, well-defined bug fix
- Request is a minor patch with obvious scope
- User has provided detailed acceptance criteria

**Tip:** When in doubt, use this skill. Clear requirements lead to better plans and fewer implementation surprises.

---

## Integration with Other Skills

### Workflow Position

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│  User Input ──▶ requirements ──▶ [design?] ──▶ classify ──▶ plan ──▶ impl  │
│                (this skill)      (if complex)              skill            │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Handoff to Design (if needed)

When `needs_design: true` or `complexity: high`, the requirements document triggers the design skill:
- Design skill reads requirements from `specs/requirements/*.md`
- Creates architecture docs at `specs/design/*.md`
- Then proceeds to classify

### Handoff to Classify

The requirements document includes a "Recommended Classification" section that helps the classify skill make faster, more accurate routing decisions.

### Handoff to Planning Skills

Planning skills (feature, bug, chore, patch, prototype) can reference the requirements document in their "Relevant Files" section to ensure alignment with defined requirements and acceptance criteria.

### Validation During Review

The review skill can reference the requirements document to validate that implementation meets the defined acceptance criteria.
