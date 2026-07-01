---
name: risk-assess
description: "Full Risk Assessment Orchestrator. Run complete risk workflow: scope → identify → assess (impact × probability) → dependencies → visualize → risk register. Use when conducting risk assessment for a project, migration, feature, or initiative. Integrates with risk-management-scenarios skill."
---

# Risk Assessment Orchestrator

Execute the complete risk assessment workflow. Chains: Scope → Identify → Assess → Dependencies → Visualize → Output. Uses **risk-management-scenarios** skill for methodology and templates.

## Workflow Overview

```
INPUT (subject: project, migration, feature, initiative, life event)
    ↓
┌──────────────────────────────────────────────────────────────────────┐
│  PHASE 1: SCOPE & CONTEXT                                             │
│  → Clarify scope, stakeholders, constraints, timeline                  │
│  → Select risk categories from domain template (Technical / Life / War)│
└──────────────────────────────────────────────────────────────────────┘
    ↓
┌──────────────────────────────────────────────────────────────────────┐
│  PHASE 2: IDENTIFY RISKS                                              │
│  → Brainstorm risks per category                                      │
│  → Use risk-assessment.md (calibration checklists, domain templates)  │
└──────────────────────────────────────────────────────────────────────┘
    ↓
┌──────────────────────────────────────────────────────────────────────┐
│  PHASE 3: ASSESS (Impact × Probability)                               │
│  → Use calibration checklists for I and P                             │
│  → Score = I×P; apply Low-Prob/High-Impact rule if I=5, P≤2           │
│  → Prioritize: 1–4 Low, 5–9 Medium, 10–15 High, 16–25 Critical       │
└──────────────────────────────────────────────────────────────────────┘
    ↓
┌──────────────────────────────────────────────────────────────────────┐
│  PHASE 3.5: RISK DEPENDENCIES (Cascading)                              │
│  → Identify triggers: R1 → R2 (R1 increases P of R2)                  │
│  → Document in Risk Dependencies table                                │
└──────────────────────────────────────────────────────────────────────┘
    ↓
┌──────────────────────────────────────────────────────────────────────┐
│  PHASE 4: VISUALIZE                                                   │
│  → Risk matrix (5×5)                                                  │
│  → risk-management-scenarios/references/visualizations.md             │
└──────────────────────────────────────────────────────────────────────┘
    ↓
┌──────────────────────────────────────────────────────────────────────┐
│  PHASE 5: OUTPUT                                                      │
│  → Risk register with Residual I×P, Mitigation Cost, Triggers          │
│  → Mitigation recommendations (concrete steps, cost estimate)         │
│  → Review Triggers section                                            │
│  → Optional: --scenario, --monte-carlo, --sensitivity, --stress, --bc-dr│
└──────────────────────────────────────────────────────────────────────┘
    ↓
┌──────────────────────────────────────────────────────────────────────┐
│  PHASE 6: VALIDATE (optional)                                         │
│  → Stakeholder review of scores; adjust if needed                      │
└──────────────────────────────────────────────────────────────────────┘
    ↓
OUTPUT: specs/risk-assess-{descriptive-name}.md
```

## Variables

- `input`: Subject to assess (required)
- `start_phase`: (optional) Resume from: `scope`, `identify`, `assess`, `dependencies`, `visualize`, `output`
- `--scope=<domain>`: (optional) Override: technical | life | war — selects category template
- `--scenario`: (optional) Add best/base/worst scenario simulation
- `--monte-carlo`: (optional) Add Monte Carlo simulation (distributions, P10/P50/P90)
- `--sensitivity`: (optional) Add sensitivity analysis (which variable matters most)
- `--stress`: (optional) Add stress test (extreme conditions)
- `--bc-dr`: (optional) Add BC/DR scenario section
- `--validate`: (optional) Add Validate phase reminder
- `--output`: (optional) Custom output path

## Instructions

### Phase 1: Scope & Context

1. Parse input — extract subject, constraints, stakeholders
2. Define scope — what is in/out of assessment
3. **Select domain template** from risk-assessment.md: Technical/Project | Life/Personal | War/Geopolitical
4. Document in output file under "## Scope"

### Phase 2: Identify Risks

1. Load `risk-management-scenarios/references/risk-assessment.md`
2. Use **domain template** categories for brainstorming
3. List each risk with clear description
4. Document in output under "## Identified Risks"

### Phase 3: Assess

1. Use **calibration checklists** (Impact, Probability) from risk-assessment.md
2. For each risk: assign Impact (1–5) and Probability (1–5)
3. Calculate Score = Impact × Probability
4. Apply **Low-Prob/High-Impact rule**: if I=5 and P≤2, mark "Review: High Impact", consider separate mitigation
5. Assign level: Low (1–4), Medium (5–9), High (10–15), Critical (16–25)
6. Sort by score descending
7. Document in table format

### Phase 3.5: Risk Dependencies

1. Identify cascading risks: which risk triggers which
2. Document in "Risk Dependencies" table: From | To | Relationship
3. Include in output under "## Risk Dependencies"

### Phase 4: Visualize

1. Build risk matrix (5×5)
2. Create heat map / summary table
3. Use formats from `risk-management-scenarios/references/visualizations.md`
4. Document in output under "## Risk Matrix"

### Phase 5: Output

1. Create risk register table: ID, Risk, Category, Impact, Prob, Score, **Residual I×P**, **Mitigation Cost**, **Triggers**, Owner, Mitigation, Status
2. Add mitigation recommendations for High/Critical with **concrete steps** and **cost estimate** (time/money)
3. Add **Review Triggers** section (key event, quarterly, mitigation completed)
4. If `--scenario`: add best/base/worst scenario section
5. If `--monte-carlo`: add Monte Carlo simulation
6. If `--sensitivity`: add sensitivity analysis
7. If `--stress`: add stress test
8. If `--bc-dr`: add BC/DR scenario section
9. Write to `specs/risk-assess-{descriptive-name}.md`

### Phase 6: Validate (optional, if --validate)

1. Add reminder: "Stakeholder review recommended for Critical/High risks"
2. Document escalation path for Critical

## Output File Structure

```markdown
# Risk Assessment — [Subject]

**Date:** YYYY-MM-DD | **Next review:** YYYY-MM-DD

## Як читати цей звіт
(коротка навігація — 1–2 речення на кожен розділ)

## Детальна розшифровка шкал
(таблиці Impact, Probability, Score з прикладами для контексту; використати calibration checklists)

## Scope
(категорії з domain template — "Що включає" / "Приклади")

## Identified Risks
(+ "Як читати")

## Assessment Table (I×P)
(+ Low-Prob/High-Impact позначки якщо є)

## Risk Dependencies (Cascading)
(From | To | Relationship)

## Risk Matrix
(+ як читати)

## Risk Register
(ID, Risk, Category, I, P, Score, Residual I×P, Mitigation Cost, Triggers, Owner, Mitigation, Status)

## Mitigation Priorities
(Конкретні кроки + орієнтовна вартість для High/Critical)

## Review Triggers
(Ключова подія | Квартальний | Mitigation завершено)

## Summary
(Level | Count | Action | "Що це означає")
(Детальні висновки)
(Ключові бар'єри)

## [Optional] Scenario Simulation / Monte Carlo / Sensitivity / Stress / BC/DR
```

## Output Quality — User-Friendly Formulations

**Звіт має бути зрозумілим без додаткового контексту.** Обов'язково включати:

1. **Розшифровка шкал** — з calibration checklists, приклади для предмета оцінки
2. **"Як читати" під кожною таблицею**
3. **Категорії** — з domain template
4. **Mitigation з конкретними кроками** — не лише "Emergency fund", а "1) Накопичити 6× місячних витрат. 2) ..."
5. **Mitigation Cost** — орієнтовна вартість (час, гроші) для High/Critical
6. **Residual I×P** — target score після mitigation
7. **Review Triggers** — коли переглядати
8. **Summary з "Що це означає"**
9. **Ключові бар'єри** — Бар'єр | Що робити | Покриває ризики

## Quick Reference — Scales

**Impact (1–5):** 1 Negligible, 2 Low, 3 Medium, 4 High, 5 Critical  
**Probability (1–5):** 1 Very Low (<1%/yr), 2 Low, 3 Medium, 4 High, 5 Very High (>70%)  
**Score = I × P** → 1–4 Low, 5–9 Medium, 10–15 High, 16–25 Critical  
**Low-Prob/High-Impact:** I=5, P≤2 → mark "Review: High Impact"

## Integration

- **risk-management-scenarios** skill provides: risk-assessment.md (calibration, domain templates, dependencies, residual, review triggers), templates
- **Command**: `/risk-assess <subject>` invokes this orchestrator
- **Resume**: If `agentic/risk-assess-{name}/state.json` exists, check `next_phase` and continue
