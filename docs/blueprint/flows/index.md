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
| 120 | [Check for drift](./cli/120-check-drift/index.md)              |           | [Zero setup drift](../product.md#goal-zero-drift), [Agent-resolvable drift](../product.md#goal-agent-resolvable) | [Group](../entities/group/index.md), [Setup config](../entities/setup-config/index.md) | reviewed |
| 130 | [Update a repository](./cli/130-update-repository/index.md)    |           | [Zero setup drift](../product.md#goal-zero-drift)    | [Group](../entities/group/index.md), [Setup config](../entities/setup-config/index.md)                          | reviewed |
| 140 | [Add a group](./cli/140-add-group/index.md)                    |           | [Fast new-repo setup](../product.md#goal-fast-setup) | [Group](../entities/group/index.md), [Setup config](../entities/setup-config/index.md)                          | reviewed |
| 150 | [Remove a group](./cli/150-remove-group/index.md)              |           | [Zero setup drift](../product.md#goal-zero-drift)    | [Group](../entities/group/index.md), [Setup config](../entities/setup-config/index.md)                          | reviewed |

### site

| #   | Flow                                  | Platforms                           | Serves goal                                                  | Entities touched | Status |
| --- | ------------------------------------- | ----------------------------------- | ------------------------------------------------------------ | ---------------- | ------ |
| 100 | [Home](./site/100-home/index.md)      | [site](./site/100-home/site.md)     | [Outside adoption](../product.md#goal-outside-adoption)      | none             | reviewed |
| 110 | [Documentation](./site/110-documentation/index.md) | [site](./site/110-documentation/site.md) | [Outside adoption](../product.md#goal-outside-adoption), [Agent-resolvable drift](../product.md#goal-agent-resolvable) | none; describes [Group](../entities/group/index.md), [Setup config](../entities/setup-config/index.md) | reviewed |

## Inter-Service Contracts

### Events

| Event | Payload contract | Producer | Consumers | Delivery semantics |
| ----- | ---------------- | -------- | --------- | ------------------ |

### Synchronous calls

| Caller → Callee | Contract | Timeout / Retry | Failure behavior |
| --------------- | -------- | --------------- | ---------------- |

## Consistency Boundaries

- None — the product has no service projects; every flow runs inside one local command or one static page, so no consistency boundary spans projects.
