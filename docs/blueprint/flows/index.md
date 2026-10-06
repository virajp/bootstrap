---
type: vwf-integration
title: Flows & Cross-Flow Contracts
description: Catalog of the product's flows plus the inter-service contracts
  and consistency boundaries no single flow owns.
status: draft
---

# Flows & Cross-Flow Contracts

## Flow catalog

### cli

| #   | Flow                                                           | Platforms | Serves goal                                          | Entities touched                                                                                                | Status |
| --- | -------------------------------------------------------------- | --------- | ---------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- | ------ |
| 110 | [Set up a repository](./cli/110-setup-repository/index.md)     |           | [Fast new-repo setup](../product.md#goal-fast-setup) | [Group](../entities/group/index.md), [Setup config](../entities/setup-config/index.md)                          | reviewed |

## Inter-Service Contracts

### Events

| Event | Payload contract | Producer | Consumers | Delivery semantics |
| ----- | ---------------- | -------- | --------- | ------------------ |

### Synchronous calls

| Caller → Callee | Contract | Timeout / Retry | Failure behavior |
| --------------- | -------- | --------------- | ---------------- |

## Consistency Boundaries

- None — the product has no service projects; every flow runs inside one local command or one static page, so no consistency boundary spans projects.
