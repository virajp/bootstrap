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

| #   | Flow                                                       | Platforms | Serves goal                                                                                             | Entities touched                                                                                                                          | Status   |
| --- | ---------------------------------------------------------- | --------- | ------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| 110 | [Set up a repository](./cli/110-setup-repository/index.md) |           | [Fast new-repo setup](../product.md#goal-fast-setup), [Zero setup drift](../product.md#goal-zero-drift) | [Tool](../entities/tool/index.md), [Tool category](../entities/tool-category/index.md), [Setup config](../entities/setup-config/index.md) | reviewed |
| 140 | [Add a tool](./cli/140-add-tool/index.md)                  |           | [Fast new-repo setup](../product.md#goal-fast-setup)                                                    | [Tool](../entities/tool/index.md), [Tool category](../entities/tool-category/index.md), [Setup config](../entities/setup-config/index.md) | reviewed |
| 150 | [Remove a tool](./cli/150-remove-tool/index.md)            |           | [Zero setup drift](../product.md#goal-zero-drift)                                                       | [Tool](../entities/tool/index.md), [Setup config](../entities/setup-config/index.md)                                                      | reviewed |
| 160 | [Select tools](./cli/160-select-tools/index.md)            |           | [Fast new-repo setup](../product.md#goal-fast-setup)                                                    | [Tool category](../entities/tool-category/index.md), [Tool](../entities/tool/index.md), [Setup config](../entities/setup-config/index.md) | reviewed |
| 170 | [Show the setup](./cli/170-show-setup/index.md)            |           | [Zero setup drift](../product.md#goal-zero-drift)                                                       | [Setup config](../entities/setup-config/index.md), [Tool](../entities/tool/index.md), [Tool category](../entities/tool-category/index.md) | reviewed |
| 180 | [Manage scopes](./cli/180-manage-scopes/index.md)          |           | [Zero setup drift](../product.md#goal-zero-drift)                                                       | [Setup config](../entities/setup-config/index.md), [Tool](../entities/tool/index.md)                                                      | reviewed |
| 190 | [Manage members](./cli/190-manage-members/index.md)        |           | [Zero setup drift](../product.md#goal-zero-drift)                                                       | [Setup config](../entities/setup-config/index.md), [Tool](../entities/tool/index.md)                                                      | reviewed |

### site

| #   | Flow                                               | Platforms                                | Serves goal                                                                                                   | Entities touched                                                                                                                                          | Status   |
| --- | -------------------------------------------------- | ---------------------------------------- | ------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| 100 | [Home](./site/100-home/index.md)                   | [site](./site/100-home/site.md)          | [Outside adoption](../product.md#goal-outside-adoption)                                                       | [Tool](../entities/tool/index.md), [Tool category](../entities/tool-category/index.md), [Setup config](../entities/setup-config/index.md)                 | reviewed |
| 110 | [Documentation](./site/110-documentation/index.md) | [site](./site/110-documentation/site.md) | [Outside adoption](../product.md#goal-outside-adoption), [Fast new-repo setup](../product.md#goal-fast-setup) | none; describes [Tool](../entities/tool/index.md), [Tool category](../entities/tool-category/index.md), [Setup config](../entities/setup-config/index.md) | reviewed |

## Inter-Service Contracts

### Events

| Event | Payload contract | Producer | Consumers | Delivery semantics |
| ----- | ---------------- | -------- | --------- | ------------------ |

### Synchronous calls

| Caller → Callee | Contract | Timeout / Retry | Failure behavior |
| --------------- | -------- | --------------- | ---------------- |

## Consistency Boundaries

- None — the product has no service projects; every flow runs inside one local
  command or one static page, so no consistency boundary spans projects.
