---
type: vwf-product
title: bootstrap — Product
description: Problem, users, success metrics, and slice priority — the outcome
  contract the blueprint serves.
status: draft # draft | reviewed | stable
timestamp: 2026-10-08
---

# bootstrap — Product

> **Source of truth for why the product exists and what "good" means.** Lives at
> `docs/blueprint/product.md`, authored by `/vwf:product` — the Phase −1
> foundation before `architecture`. Code- and stack-independent: no technology,
> project, or screen names — those belong to the registry and the entity docs.
> Every flow's Purpose links the goal(s) it serves (entities trace to goals
> through the flows that use them); the blueprint-reviewer flags surfaces that
> trace to no goal.

## Problem

A developer who runs many repositories sets up each one's development tooling —
the formatter, the commit gates, the secret and vulnerability scanners, the task
library, the ignore files, the editor defaults — by copying configuration from
another repository and editing it by hand. That is slow on day zero, and it
keeps costing afterwards: copies go stale, carry the source repository's names
and paths, and drift apart from each other, so time goes into fixing setup
instead of building the product.

Day-zero setup and ongoing drift are one problem: a repository's setup should
come from one shared source, both when the repository is created and for the
rest of its life. Each change to the setup is an entry in the repository's
version history, so the owner can go back to any earlier state.

**Why now:** the number of repositories keeps growing, and the current way of
producing this setup — embedded inside another tool's workflow — has become too
complex to maintain.

## Target users

| Persona           | Who they are                                                                                          | Core need                                                                                            |
| ----------------- | ----------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| Repo owner        | A developer maintaining many personal and work repositories                                           | Every repository set up the same way from one source, with no hand-copying and no hand-fixing        |
| AI coding agent   | An agent working in a repository on the owner's behalf                                                | Set up a repository, and add or remove a tool, without prompts, and read one machine-readable result |
| Outside developer | A developer outside the owner's repositories who adopts the published tool for their own repositories | A sensible, opinionated setup without having to design one                                           |

## Goals & success metrics

### Zero setup drift {#goal-zero-drift}

- Outcome: every active repository's setup matches the shared source.
- Metric: share of the owner's active repositories where a re-run of the setup
  with the latest release leaves no uncommitted change — target 100% within 3
  months of the 1.0 release
- Measured via: external the owner's re-run script over the active repositories
- Re-evaluate if: share below 50% by 6 months after the 1.0 release → re-scope

### Fast new-repo setup {#goal-fast-setup}

- Outcome: a new repository is ready to build in minutes.
- Metric: time from an empty repository to all gates passing — target under 5
  minutes at the 1.0 release
- Measured via: external timed runs in a scratch repository

### Outside adoption {#goal-outside-adoption}

- Outcome: developers outside the owner's repositories adopt the tool.
- Metric: weekly installs of the published tool — target 100 per week within 6
  months of the 1.0 release
- Measured via: external public install statistics

## Slice priority

| Rank | Slice (flow / entity) | Serves goal                                                                   | Validates                                                  | Why now                                                                                                                    |
| ---- | --------------------- | ----------------------------------------------------------------------------- | ---------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| 1    | Set up a repository   | [Fast new-repo setup](#goal-fast-setup), [Zero setup drift](#goal-zero-drift) | [Re-running replaces committed edits](#risks--assumptions) | The core path — rendering the shared source into a repository, first time and every re-run — that every other slice reuses |
| 2    | Add a tool            | [Fast new-repo setup](#goal-fast-setup)                                       | —                                                          | Lets a repository take on a tool later, or replace one, without starting over                                              |
| 3    | Remove a tool         | [Zero setup drift](#goal-zero-drift)                                          | —                                                          | Lets a repository drop a tool it no longer wants, without leftover files                                                   |
| 4    | Select tools          | [Fast new-repo setup](#goal-fast-setup)                                       | —                                                          | Shows every tool by category in one view, so the owner adds and removes tools in one place                                 |
| 5    | Show the setup        | [Zero setup drift](#goal-zero-drift)                                          | —                                                          | Lets an owner or agent read the recorded setup, or one field of it, without opening the file                               |
| 6    | Documentation         | [Outside adoption](#goal-outside-adoption)                                    | —                                                          | What an outside developer reads before adopting; ships with 1.0                                                            |

```mermaid
flowchart LR
    A[Set up a repository] --> B[Add a tool]
    B --> C[Remove a tool]
    C --> D[Select tools]
    D --> F[Show the setup]
    F --> E[Documentation]
```

Rank 1 validates only the re-run assumption. The riskiest assumption — one
shared source fits every repository — is validated by its cheaper method
(rendering against the existing repositories) before slice 1 is built.

## Non-goals

- **No application scaffolding.** It sets up tooling and hygiene only; it never
  generates application code, frameworks or project structure.
- **No automatic commits.** It writes files, and the owner commits them; the
  version history holds only the owner's commits.
- **Never overwrites uncommitted work.** A file it would change that has
  uncommitted changes stops the whole run, and nothing is written.
- **No language, cloud or deploy setup in 1.0.** Language toolchains and cloud
  or deploy configuration are left out, possibly to become add-ons later. A
  runtime that a setup tool itself needs is installed only as that tool's
  dependency.
- **No hosted service.** It runs locally and in continuous integration only — no
  accounts, no server, no telemetry.

## Risks & assumptions

| Assumption                                                                                                        | Risk if wrong                                                                | Validation method                                                    | Status   | Evidence |
| ----------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- | -------------------------------------------------------------------- | -------- | -------- |
| One shared source fits every repository, with differences confined to values and opt-in tools                     | Repositories fork the source to fit, and drift returns                       | usage-data                                                           | untested | —        |
| One tool category list fits every repository, with at most one tool in each single-tool category                  | Repositories need two tools where the source allows one, and fork the source | usage-data                                                           | untested | —        |
| Owners accept that a re-run replaces committed edits to the files it owns, because the version history keeps them | Owners stop re-running, and existing repositories drift indefinitely         | slice:set-up-a-repository                                            | untested | —        |
| Outside developers want an opinionated setup rather than designing their own                                      | Adoption stays near zero                                                     | accepted-risk — a side benefit that does not block the owner's goals | untested | —        |

The first two rows are validated by rendering the shared source against the
seven repositories that already carry this setup, and counting differences that
are not values and tool categories that need more than one tool.
